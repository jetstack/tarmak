package rpc

import (
	"fmt"
	"io"
	"log"
	"net"
	"net/rpc"
	"os"
	"os/signal"
	"syscall"
)

const (
	serverName = "Tarmak"
	socketName = "tarmak.sock"
)

type tarmakRPC struct{}

func (i *tarmakRPC) BastionStatus(args string, reply *string) error {

	fmt.Printf("BastionStatus called\n")

	// TODO: actually check if bastion is up
	*reply = "running"

	return nil
}

// Start starts an RPC server to serve requests from
// the container
func Start() error {

	fmt.Printf("starting %s RPC server\n", serverName)
	ln, err := net.Listen("unix", socketName)
	if err != nil {
		return fmt.Errorf("unable to listen on socket %s: %s", socketName, err)
	}

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, os.Interrupt, syscall.SIGTERM)
	go func(ln net.Listener, c chan os.Signal) {
		sig := <-c
		log.Printf("caught signal %s: shutting down.", sig)
		ln.Close()
		//os.Exit(0)
	}(ln, sigc)

	for {
		fd, err := ln.Accept()
		if err != nil {
			fmt.Printf("error accepting RPC request: %s", err)
		}

		go accept(fd)
	}
}

func accept(conn net.Conn) {

	s := rpc.NewServer()
	s.RegisterName(serverName, &tarmakRPC{})

	fmt.Printf("Connection made\n")

	s.ServeConn(struct {
		io.Reader
		io.Writer
		io.Closer
	}{conn, conn, conn})

}
