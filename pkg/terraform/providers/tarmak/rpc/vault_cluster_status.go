// Copyright Jetstack Ltd. See LICENSE for details.
package rpc

import (
	"fmt"

	"github.com/jetstack/tarmak/pkg/tarmak/stack"
	"github.com/jetstack/vault-helper/pkg/kubernetes"
)

var (
	VaultClusterStatusCall     = fmt.Sprintf("%s.VaultClusterStatus", RPCName)
	VaultClusterInitStatusCall = fmt.Sprintf("%s.VaultClusterInitStatus", RPCName)
)

type VaultClusterStatusArgs struct {
	VaultInternalFQDNs []string
	VaultCA            string
	VaultKMSKeyID      string
	VaultUnsealKeyName string
}

type VaultClusterStatusReply struct {
	Status string
}

func (r *tarmakRPC) VaultClusterStatus(args *VaultClusterStatusArgs, result *VaultClusterStatusReply) error {
	r.tarmak.Log().Debug("received rpc vault cluster status")

	// TODO: if destroying cluster just return unknown here

	///CUSTOM
	vaultStack := r.tarmak.Cluster().Environment().VaultStack()
	//vaultStack := r.tarmak.Cluster().Environment().Stack(tarmakv1alpha1.StackNameVault)

	vaultStackReal, ok := vaultStack.(*stack.VaultStack)
	if !ok {
		err := fmt.Errorf("unexpected type for vault stack: %T", vaultStack)
		r.tarmak.Log().Error(err)
		return err
	}
	err := vaultStackReal.VerifyVaultInitFromFQDNs(args.VaultInternalFQDNs, args.VaultCA, args.VaultKMSKeyID, args.VaultUnsealKeyName)
	if err != nil {
		err = fmt.Errorf("failed to initialise vault cluster: %s", err)
		r.tarmak.Log().Error(err)
		return err
	}
	///CUSTOM

	//initVaultClient

	// load outputs from terraform
	r.tarmak.Cluster().Environment().Tarmak().Terraform().Output(vaultStack)

	vaultTunnel, err := vaultStackReal.VaultTunnelFromFQDNs(args.VaultInternalFQDNs, args.VaultCA)
	if err != nil {
		err = fmt.Errorf("failed to create vault tunnel: %s", err)
		r.tarmak.Log().Error(err)
		return err
	}
	defer vaultTunnel.Stop()

	vaultClient := vaultTunnel.VaultClient()

	vaultRootToken, err := r.tarmak.Cluster().Environment().VaultRootToken()
	if err != nil {
		err = fmt.Errorf("failed to retrieve vault root token: %s", err)
		r.tarmak.Log().Error(err)
		return err
	}

	vaultClient.SetToken(vaultRootToken)

	//initVaultClient

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

	//initVaultClient
	vaultStack := r.tarmak.Cluster().Environment().VaultStack()
	//vaultStack := r.tarmak.Cluster().Environment().Stack(tarmakv1alpha1.StackNameVault)

	vaultStackReal, ok := vaultStack.(*stack.VaultStack)
	if !ok {
		err := fmt.Errorf("unexpected type for vault stack: %T", vaultStack)
		r.tarmak.Log().Error(err)
		return err
	}

	// load outputs from terraform
	r.tarmak.Cluster().Environment().Tarmak().Terraform().Output(vaultStack)

	vaultTunnel, err := vaultStackReal.VaultTunnelFromFQDNs(args.VaultInternalFQDNs, args.VaultCA)
	if err != nil {
		err = fmt.Errorf("failed to create vault tunnel: %s", err)
		r.tarmak.Log().Error(err)
		return err
	}
	defer vaultTunnel.Stop()

	vaultClient := vaultTunnel.VaultClient()

	vaultRootToken, err := r.tarmak.Cluster().Environment().VaultRootToken()
	if err != nil {
		err = fmt.Errorf("failed to retrieve vault root token: %s", err)
		r.tarmak.Log().Error(err)
		return err
	}

	vaultClient.SetToken(vaultRootToken)

	//initVaultClient

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
