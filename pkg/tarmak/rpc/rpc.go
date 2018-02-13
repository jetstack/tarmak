package rpc

import (
	"fmt"
	"io"
	"net"
	"net/rpc"
	"time"

	tarmakv1alpha1 "github.com/jetstack/tarmak/pkg/apis/tarmak/v1alpha1"
	"github.com/jetstack/tarmak/pkg/tarmak"
	"github.com/jetstack/tarmak/pkg/tarmak/cluster"
	"github.com/jetstack/tarmak/pkg/tarmak/stack"
)

const (
	serverName        = "Tarmak"
	socketName        = "tarmak.sock"
	onFailureWaitTime = 10 * time.Second
)

type tarmakRPC struct {
	tarmak *tarmak.Tarmak
}

func (i *tarmakRPC) BastionInstanceStatus(args [2]string, reply *string) error {

	fmt.Printf("BastionInstanceStatus called\n")

	//hostname := args[0]
	//username := args[1]

	t := i.tarmak
	c, err := cluster.NewFromConfig(t.Environment(), t.Cluster().Config())
	if err != nil {
		*reply = "down"
		return fmt.Errorf("failed to retreive cluster: %s", err)
	}

	for {
		tunnel := c.Environment().WingTunnel()
		err = tunnel.Start()
		if err != nil {
			time.Sleep(onFailureWaitTime)
			continue
		}

		*reply = "up"
		return nil
	}

}

func (i *tarmakRPC) VaultClusterStatus(instances []string, reply *string) error {

	fmt.Printf("VaultClusterStatus called\n")

	t := i.tarmak

	// build vault stack
	// TODO: Look at ensureVaultSetup in kubernetes_stack.go for a better way of doing this
	s := &stack.Stack{}
	s.SetCluster(t.Cluster())
	s.SetLog(t.Cluster().Log().WithField("stack", tarmakv1alpha1.StackNameVault))

	v, err := stack.NewVaultStack(s)
	if err != nil {
		return fmt.Errorf("error while creating vault stack: %s", err)
	}

	output, err := t.Terraform().Output(v)
	if err != nil {
		return fmt.Errorf("error while getting terraform output: %s", err)
	}
	v.SetOutput(output)

	for {
		err = v.VerifyVaultInitForFQDNs(instances)
		if err != nil {
			fmt.Printf("failed to connect to vault: %s", err)
			time.Sleep(onFailureWaitTime)
			continue
		}

		*reply = "up"
		return nil
	}

}

func (i *tarmakRPC) VaultInstanceRoleStatus(args [2]string, reply *string) error {
	fmt.Printf("VaultInstanceRoleStatus called\n")

	//vaultClusterName := args[0]
	roleName := args[1]

	t := i.tarmak
	clusterStacks := t.Cluster().Stacks()

	for {
		for _, clusterStack := range clusterStacks {
			if clusterStack.Name() == tarmakv1alpha1.StackNameKubernetes {

				// get real kubernetes stack
				kubernetesStack, ok := clusterStack.(*stack.KubernetesStack)
				if !ok {
					return fmt.Errorf("unexpected type for kubernetes stack: %T", clusterStack)
				}

				// retrieve init tokens
				err := kubernetesStack.EnsureVaultSetup()
				if err != nil {
					return fmt.Errorf("error ensuring vault setup: %s", err)
				}

				// test existence of init token for role
				initTokens := kubernetesStack.InitTokens()
				_, ok = initTokens[fmt.Sprintf("vault_init_token_%s", roleName)]

				if ok {
					*reply = "up"
					return nil
				}
			}
		}
		fmt.Printf("failed to retrieve init token for %s", roleName)
		time.Sleep(onFailureWaitTime)
		continue
	}
}

// Start starts an RPC server to serve requests from
// the container
func Start(t *tarmak.Tarmak) error {

	fmt.Printf("starting %s RPC server\n", serverName)
	ln, err := net.Listen("unix", socketName)
	if err != nil {
		return fmt.Errorf("unable to listen on socket %s: %s", socketName, err)
	}

	for {
		fd, err := ln.Accept()
		if err != nil {
			fmt.Printf("error accepting RPC request: %s", err)
		}

		go accept(fd, t)
	}
}

func accept(conn net.Conn, tarmak *tarmak.Tarmak) {

	tarmakRPC := tarmakRPC{tarmak: tarmak}

	s := rpc.NewServer()
	s.RegisterName(serverName, &tarmakRPC)

	fmt.Printf("Connection made\n")

	s.ServeConn(struct {
		io.Reader
		io.Writer
		io.Closer
	}{conn, conn, conn})

}
