// Copyright Jetstack Ltd. See LICENSE for details.
package rpc

import (
	"fmt"

	"github.com/jetstack/tarmak/pkg/tarmak/stack"
	"github.com/jetstack/vault-helper/pkg/kubernetes"

	vault "github.com/hashicorp/vault/api"
)

var (
	VaultClusterStatusCall     = fmt.Sprintf("%s.VaultClusterStatus", RPCName)
	VaultClusterInitStatusCall = fmt.Sprintf("%s.VaultClusterInitStatus", RPCName)
)

type VaultClusterStatusArgs struct {
	VaultInternalFQDNs []string
}

type VaultClusterStatusReply struct {
	Status string
}

func (r *tarmakRPC) VaultClusterStatus(args *VaultClusterStatusArgs, result *VaultClusterStatusReply) error {
	r.tarmak.Log().Debug("received rpc vault cluster status")

	// TODO: if destroying cluster just return unknown here

	vaultClient, err := initVaultClient(r)
	if err != nil {
		return err
	}

	k := kubernetes.New(vaultClient, r.tarmak.Log())
	k.SetClusterID(r.tarmak.Cluster().ClusterName())

	if err := k.Ensure(); err != nil {
		err = fmt.Errorf("vault cluster is not ready: %s", err)
		r.tarmak.Log().Error(err)
		return err
	}

	result.Status = "ready"
	return nil
}

func (r *tarmakRPC) VaultClusterInitStatus(args *VaultClusterStatusArgs, result *VaultClusterStatusReply) error {
	r.tarmak.Log().Debug("received rpc vault cluster status")

	vaultClient, err := initVaultClient(r)
	if err != nil {
		return err
	}

	up, err := vaultClient.Sys().InitStatus()
	if err != nil {
		err = fmt.Errorf("failed to retrieve init status: %s", err)
		r.tarmak.Log().Error(err)
		return err
	}

	if !up {
		err = fmt.Errorf("vault cluster has not been initialised: %s", err)
		r.tarmak.Log().Error(err)
		return err
	}

	result.Status = "ready"
	return nil
}

func initVaultClient(r *tarmakRPC) (*vault.Client, error) {
	//vaultStack := r.tarmak.Cluster().Environment().Stack(tarmakv1alpha1.StackNameVault)
	vaultStack := r.tarmak.Cluster().Environment().VaultStack()

	// load outputs from terraform
	r.tarmak.Cluster().Environment().Tarmak().Terraform().Output(vaultStack)

	vaultStackReal, ok := vaultStack.(*stack.VaultStack)
	if !ok {
		err := fmt.Errorf("unexpected type for vault stack: %T", vaultStack)
		r.tarmak.Log().Error(err)
		return nil, err
	}

	vaultTunnel, err := vaultStackReal.VaultTunnel()
	if err != nil {
		err = fmt.Errorf("failed to create vault tunnel: %s", err)
		r.tarmak.Log().Error(err)
		return nil, err
	}
	defer vaultTunnel.Stop()

	vaultClient := vaultTunnel.VaultClient()

	vaultRootToken, err := r.tarmak.Cluster().Environment().VaultRootToken()
	if err != nil {
		err = fmt.Errorf("failed to retrieve vault root token: %s", err)
		r.tarmak.Log().Error(err)
		return nil, err
	}

	vaultClient.SetToken(vaultRootToken)
	return vaultClient, nil
}
