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

func (i *tarmakRPC) BastionInstanceStatus(hostname string, reply *string) error {

	fmt.Printf("BastionInstanceStatus called\n")

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
	s := &stack.Stack{}
	s.SetCluster(t.Cluster())
	s.SetLog(t.Cluster().Log().WithField("stack", tarmakv1alpha1.StackNameVault))

	v, err := stack.NewVaultStack(s)
	if err != nil {
		return fmt.Errorf("error while getting vault stack: %s", err)
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

/*func (i *tarmakRPC) VaultInstanceRoleStatus(args string, reply *string) error {
	fmt.Printf("VaultInstanceRoleStatus called\n")

	t := i.tarmak
}*/

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
