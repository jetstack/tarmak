// Copyright Jetstack Ltd. See LICENSE for details.
package rpc

import (
	"fmt"

	"github.com/jetstack/vault-helper/pkg/kubernetes"

	cluster "github.com/jetstack/tarmak/pkg/apis/cluster/v1alpha1"
)

var (
	VaultInstanceRoleCreate = fmt.Sprintf("%s.VaultInstanceRoleCreate", RPCName)
	VaultInstanceRoleRead   = fmt.Sprintf("%s.VaultInstanceRoleRead", RPCName)
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

func (r *tarmakRPC) VaultInstanceRoleCreate(args *VaultInstanceRoleArgs, result *VaultInstanceRoleReply) error {
	r.tarmak.Log().Debug("received rpc vault instance role create")

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

	initToken, err := k.NewInitToken(roleName)
	if err != nil {
		return fmt.Errorf("could not get init token for role %s: %s", roleName, err)
	}

	err = initToken.Ensure()
	if err != nil {
		return fmt.Errorf("could not ensure init token for role %s: %s", roleName, err)
	}

	initTokenString, err := initToken.InitToken()
	if err != nil {
		return fmt.Errorf("could not retrieve init token for role %s: %s", roleName, err)
	}

	result.InitToken = initTokenString

	r.tarmak.Log().Debug(roleName, " init token ", initTokenString)

	return nil
}

func (r *tarmakRPC) VaultInstanceRoleRead(args *VaultInstanceRoleArgs, result *VaultInstanceRoleReply) error {
	r.tarmak.Log().Debug("received rpc vault instance role read")

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

	initToken, err := k.NewInitToken(roleName)
	if err != nil {
		return fmt.Errorf("could not get init token for role %s: %s", roleName, err)
	}

	initTokenString, err := initToken.InitToken()
	if err != nil {
		return fmt.Errorf("could not retrieve init token for role %s: %s", roleName, err)
	}

	result.InitToken = initTokenString

	r.tarmak.Log().Debug(roleName, " init token ", initTokenString)

	return nil
}
