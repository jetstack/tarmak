// Copyright Jetstack Ltd. See LICENSE for details.
package rpc

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/rpc"
	"time"

	"github.com/cenkalti/backoff"

	tarmakv1alpha1 "github.com/jetstack/tarmak/pkg/apis/tarmak/v1alpha1"
	"github.com/jetstack/tarmak/pkg/tarmak"
	"github.com/jetstack/tarmak/pkg/tarmak/cluster"
	"github.com/jetstack/tarmak/pkg/tarmak/stack"
	"github.com/jetstack/tarmak/pkg/tarmak/utils"
)

const (
	serverName = "Tarmak"
	socketName = "tarmak.sock"
)

type tarmakRPC struct {
	tarmak *tarmak.Tarmak
}

// Start starts an RPC server to serve requests from
// the container
func Start(t *tarmak.Tarmak) error {
	t.Log().Infof("Starting %s RPC server", serverName)

	stopCh := utils.BasicSignalHandler(t.Log())

	ln, err := net.Listen("unix", socketName)
	if err != nil {
		return fmt.Errorf("unable to listen on socket %s: %s", socketName, err)
	}

	go func() {
		<-stopCh
		if err := ln.Close(); err != nil {
			t.Log().Errorf("failed to close rpc server: %v", err)
		}
	}()

	for {
		select {
		case <-stopCh:
			return nil

		default:
			fd, err := ln.Accept()
			if err != nil {
				continue
			}

			go accept(fd, t)
		}
	}

	return nil
}

func accept(conn net.Conn, tarmak *tarmak.Tarmak) {
	tarmakRPC := tarmakRPC{tarmak: tarmak}

	s := rpc.NewServer()
	s.RegisterName(serverName, &tarmakRPC)

	tarmak.Log().Debugf("Connection made.")

	s.ServeConn(struct {
		io.Reader
		io.Writer
		io.Closer
	}{conn, conn, conn})
}

func (i *tarmakRPC) BastionInstanceStatus(args [2]string, reply *string) error {

	i.tarmak.Log().Debugf("BastionInstanceStatus called.")

	//hostname := args[0]
	//username := args[1]

	t := i.tarmak
	c, err := cluster.NewFromConfig(t.Environment(), t.Cluster().Config())
	if err != nil {
		*reply = "down"
		return fmt.Errorf("failed to retreive cluster: %s", err)
	}

	b := i.newBackOff()
	wingTunnel := func() error {
		err := c.Environment().WingTunnel().Start()
		if err != nil {
			return err
		}

		*reply = "up"
		return nil
	}

	if err := backoff.Retry(wingTunnel, b); err != nil {
		return fmt.Errorf("unable to retrieve bastion status: %v", err)
	}

	return nil
}

func (i *tarmakRPC) VaultClusterStatus(instances []string, reply *string) error {

	i.tarmak.Log().Debugf("VaultClusterStatus called.")

	t := i.tarmak

	vaultStack := t.Cluster().Environment().VaultStack()

	// load outputs from terraform
	t.Cluster().Environment().Tarmak().Terraform().Output(vaultStack)

	vaultStackReal, ok := vaultStack.(*stack.VaultStack)
	if !ok {
		return fmt.Errorf("unexpected type for vault stack: %T", vaultStack)
	}

	b := i.newBackOff()
	verifyVault := func() error {
		err := vaultStackReal.VerifyVaultInitForFQDNs(instances)
		if err != nil {
			return err
		}

		*reply = "up"
		return nil
	}

	if err := backoff.Retry(verifyVault, b); err != nil {
		return fmt.Errorf("unable to verify vault cluster status: %v", err)
	}

	return nil
}

func (i *tarmakRPC) VaultInstanceRoleStatus(args string, reply *string) error {
	i.tarmak.Log().Debugf("VaultInstanceRoleStatus called.")

	b := i.newBackOff()
	vaultRoleStatus := func() error {
		for {
			for _, clusterStack := range i.tarmak.Cluster().Stacks() {
				if clusterStack.Name() == tarmakv1alpha1.StackNameKubernetes {

					// get real kubernetes stack
					kubernetesStack, ok := clusterStack.(*stack.KubernetesStack)
					if !ok {
						return fmt.Errorf("unexpected type for kubernetes stack: %T", clusterStack)
					}

					// attempt to retrieve init tokens
					err := kubernetesStack.EnsureVaultSetup()
					if err != nil {
						return fmt.Errorf("error ensuring vault setup: %s", err)
					}

					// test existence of init token for role
					initTokens := kubernetesStack.InitTokens()
					initToken, ok := initTokens[fmt.Sprintf("vault_init_token_%s", args)]

					if ok {
						*reply = initToken.(string)
						return nil
					}
				}
			}

			return fmt.Errorf("failed to retrieve init token for role %s", args)
		}
	}

	if err := backoff.Retry(vaultRoleStatus, b); err != nil {
		return fmt.Errorf("failed to retrive vault instance role status: %v", err)
	}

	return nil
}

func (i *tarmakRPC) Handshake(args string, reply *string) error {
	*reply = "Hello from the other side!"

	return nil
}

func (i *tarmakRPC) newBackOff() backoff.BackOffContext {
	expBackoff := backoff.NewExponentialBackOff()
	expBackoff.InitialInterval = time.Second
	expBackoff.MaxElapsedTime = time.Minute

	return backoff.WithContext(expBackoff, context.Background())
}
