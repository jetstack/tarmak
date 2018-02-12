package rpc

import (
	"fmt"
	"io"
	"net"
	"net/rpc"
	"time"

	"github.com/jetstack/tarmak/pkg/tarmak"
	"github.com/jetstack/tarmak/pkg/tarmak/cluster"
)

const (
	serverName        = "Tarmak"
	socketName        = "tarmak.sock"
	onFailureWaitTime = 10 * time.Second
)

type tarmakRPC struct {
	tarmak *tarmak.Tarmak
}

func (i *tarmakRPC) BastionStatus(args string, reply *string) error {

	fmt.Printf("BastionStatus called\n")

	t := i.tarmak
	c, err := cluster.NewFromConfig(t.Environment(), t.Cluster().Config())
	if err != nil {
		*reply = "down"
		return fmt.Errorf("failed to retreive cluster: %s", err)
	}

	for {
		_, err = c.WingInstanceClient()
		if err != nil {
			time.Sleep(onFailureWaitTime)
			//*reply = "down"
			//return fmt.Errorf("failed to connect to wing API on bastion") //: %s"), err)
			continue
		}

		*reply = "up"
		return nil
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
