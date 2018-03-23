// Copyright Jetstack Ltd. See LICENSE for details.
package rpc

import (
	"fmt"
	"time"

	"github.com/jetstack/tarmak/pkg/tarmak/stack"
	"github.com/jetstack/vault-helper/pkg/kubernetes"
)

var (
	VaultInstanceRole = fmt.Sprintf("%s.VaultInstanceRole", RPCName)
)

type VaultInstanceRoleArgs struct {
	VaultClusterName   string
	RoleName           string
	VaultInternalFQDNs []string
	VaultCA            string
	Create             bool
}

type VaultInstanceRoleReply struct {
	InitToken string
}

func (r *tarmakRPC) VaultInstanceRole(args *VaultInstanceRoleArgs, result *VaultInstanceRoleReply) error {
	r.tarmak.Log().Debug("received rpc vault instance role")

	roleName := args.RoleName

	// TODO: if destroying cluster just return unknown here

	vaultStack := r.tarmak.Cluster().Environment().VaultStack()

	vaultStackReal, ok := vaultStack.(*stack.VaultStack)
	if !ok {
		err := fmt.Errorf("unexpected type for vault stack: %T", vaultStack)
		r.tarmak.Log().Error(err)
		return err
	}
	vaultTunnel, err := vaultStackReal.VaultTunnelFromFQDNs(args.VaultInternalFQDNs, args.VaultCA)
	if err != nil {
		err := fmt.Errorf("failed to create vault tunnel: %s", err)
		r.tarmak.Log().Error(err)
		return err
	}
	defer vaultTunnel.Stop()

	vaultClient := vaultTunnel.VaultClient()

	vaultRootToken, err := r.tarmak.Cluster().Environment().VaultRootToken()
	if err != nil {
		err := fmt.Errorf("failed to retrieve root token: %s", err)
		r.tarmak.Log().Error(err)
		return err
	}

	vaultClient.SetToken(vaultRootToken)

	k := kubernetes.New(vaultClient, r.tarmak.Log())
	k.SetClusterID(r.tarmak.Cluster().ClusterName())

	token := ""
	for i := 1; i <= Retries; i++ {
		if args.Create {
			token, err = k.EnsureInitToken(roleName)
			if err == nil {
				break
			}
		} else {
			token, err = k.GetInitToken(roleName)
			if err == nil {
				break
			}
		}
		time.Sleep(time.Second)
	}
	if err != nil {
		return fmt.Errorf("could not get init token for role %s: %s", roleName, err)
	}

	result.InitToken = token
	return nil
}
