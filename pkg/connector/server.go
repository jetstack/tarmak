// Copyright Jetstack Ltd. See LICENSE for details.
package connector

import (
	"io"
	"net/rpc"
	"os"
	"time"

	"github.com/hashicorp/go-multierror"
)

type ConnectorRPC struct{}

type connectorCloser struct {
	*os.Process
}

type multiCloser struct {
	closers []io.Closer
}

func (c *Connector) StartServer() {
	server := rpc.NewServer()
	server.RegisterName("Connector", c.connRPC)
	server.ServeConn(struct {
		io.Reader
		io.Writer
		io.Closer
	}{
		os.Stdin,
		os.Stdout,
		multiCloser{
			[]io.Closer{
				os.Stdout,
				os.Stdin,
				connectorCloser{},
			},
		},
	})
}

func (c *ConnectorRPC) Hello(args int, reply *int) error {
	*reply = args + args

	return nil
}

func (mc multiCloser) Close() error {
	var result *multierror.Error

	for _, c := range mc.closers {
		if err := c.Close(); err != nil {
			result = multierror.Append(result, err)
		}
	}

	return result.ErrorOrNil()
}

func (cc connectorCloser) Close() error {
	if cc.Process == nil {
		os.Exit(0)
		return nil
	}

	c := make(chan error)
	go func() {
		_, err := cc.Process.Wait()
		c <- err
	}()

	if err := cc.Process.Signal(os.Interrupt); err != nil {
		return err
	}

	select {
	case err := <-c:
		return err
	case <-time.After(1 * time.Second):
		return cc.Process.Kill()
	}
}
