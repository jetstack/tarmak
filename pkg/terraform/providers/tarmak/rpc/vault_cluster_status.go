// Copyright Jetstack Ltd. See LICENSE for details.
package rpc

import (
	"fmt"
	"time"

	"github.com/jetstack/vault-helper/pkg/kubernetes"

	cluster "github.com/jetstack/tarmak/pkg/apis/cluster/v1alpha1"
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

	if r.tarmak.Cluster().GetState() == cluster.StateDestroy {
		result.Status = "unknown"
		return nil
	}

	vault := r.cluster.Environment().Vault()

	err := vault.VerifyInitFromFQDNs(args.VaultInternalFQDNs, args.VaultCA, args.VaultKMSKeyID, args.VaultUnsealKeyName)
	if err != nil {
		err = fmt.Errorf("failed to initialise vault cluster: %s", err)
		r.tarmak.Log().Error(err)
		return err
	}

	vaultTunnel, err := vault.TunnelFromFQDNs(args.VaultInternalFQDNs, args.VaultCA)
	if err != nil {
		err = fmt.Errorf("failed to create vault tunnel: %s", err)
		r.tarmak.Log().Error(err)
		return err
	}
	defer vaultTunnel.Stop()

	vaultClient := vaultTunnel.VaultClient()

	vaultRootToken, err := vault.RootToken()
	if err != nil {
		err = fmt.Errorf("failed to retrieve vault root token: %s", err)
		r.tarmak.Log().Error(err)
		return err
	}

	vaultClient.SetToken(vaultRootToken)

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

	if r.tarmak.Cluster().GetState() == cluster.StateDestroy {
		result.Status = "unknown"
		return nil
	}

	vault := r.cluster.Environment().Vault()

	vaultTunnel, err := vault.TunnelFromFQDNs(args.VaultInternalFQDNs, args.VaultCA)
	if err != nil {
		err = fmt.Errorf("failed to create vault tunnel: %s", err)
		r.tarmak.Log().Error(err)
		return err
	}
	defer vaultTunnel.Stop()

	vaultClient := vaultTunnel.VaultClient()

	vaultRootToken, err := vault.RootToken()
	if err != nil {
		err = fmt.Errorf("failed to retrieve vault root token: %s", err)
		r.tarmak.Log().Error(err)
		return err
	}

	vaultClient.SetToken(vaultRootToken)

	up := false
	err = nil
	for i := 1; i <= Retries; i++ {
		up, err = vaultClient.Sys().InitStatus()
		if err != nil {
			time.Sleep(time.Second)
			continue
		}
		break
	}
	if err != nil {
		err = fmt.Errorf("failed to retrieve init status: %s", err)
		r.tarmak.Log().Error(err)
		return err
	}
	if !up {
		err = fmt.Errorf("failed to initialised vault cluster")
		r.tarmak.Log().Error(err)
		return err
	}

	result.Status = "ready"
	return nil
}
