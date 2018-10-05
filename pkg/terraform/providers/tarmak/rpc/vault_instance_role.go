// Copyright Jetstack Ltd. See LICENSE for details.
package rpc

import (
	"fmt"

	"github.com/jetstack/vault-helper/pkg/kubernetes"

	cluster "github.com/jetstack/tarmak/pkg/apis/cluster/v1alpha1"
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

	if r.tarmak.Cluster().GetState() == cluster.StateDestroy {
		result.InitToken = ""
		return nil
	}

	roleName := args.RoleName

	vault := r.cluster.Environment().Vault()
	vaultTunnel, err := vault.TunnelFromFQDNs(args.VaultInternalFQDNs, args.VaultCA)
	if err != nil {
		err := fmt.Errorf("failed to create vault tunnel: %s", err)
		r.tarmak.Log().Error(err)
		return err
	}
	defer vaultTunnel.Stop()

	vaultClient := vaultTunnel.VaultClient()

	vaultRootToken, err := vault.RootToken()
	if err != nil {
		err := fmt.Errorf("failed to retrieve root token: %s", err)
		r.tarmak.Log().Error(err)
		return err
	}

	vaultClient.SetToken(vaultRootToken)

	k := kubernetes.New(vaultClient, r.tarmak.Log())
	k.SetClusterID(r.tarmak.Cluster().ClusterName())

	changesNeeded, err := k.EnsureDryRun()
	if err != nil {
		err = fmt.Errorf("vault cluster is not ready: %s", err)
		r.tarmak.Log().Error(err)
		return err
	}
	if changesNeeded {
		if err := k.Ensure(); err != nil {
			err = fmt.Errorf("vault cluster is not ready: %s", err)
			r.tarmak.Log().Error(err)
			return err
		}
	}

	initTokens := k.InitTokens()
	initToken, ok := initTokens[roleName]
	if !ok {
		return fmt.Errorf("could not get init token for role %s: %s", roleName, err)
	}

	result.InitToken = initToken
	return nil
}
