// Copyright Jetstack Ltd. See LICENSE for details.
package connector

import (
	"bufio"
	"fmt"
	"net"
	//"github.com/hashicorp/go-multierror"
)

const (
	providerSocket = "provider-tarmak.sock"
	serverName     = "tarmak-connector"
)

func (c *Connector) CloseServer() error {
	if err := c.server.Close(); err != nil {
		return fmt.Errorf("failed to close connector server: %v", err)
	}

	return nil
}

func (c *Connector) StartServer() error {
	ln, err := net.Listen("unix", providerSocket)
	if err != nil {
		return fmt.Errorf("unable to listen to provider socket: %v", err)
	}

	c.server, err = ln.Accept()
	if err != nil {
		return fmt.Errorf("unable to accept from provider socket: %v", err)
	}

	return nil
}

func (c *Connector) ForwardConnection(buf []byte) error {
	writer := bufio.NewWriter(c.server)

	for _, b := range buf {
		if _, err := writer.Write([]byte{b}); err != nil {
			return fmt.Errorf("failed to send byte to provider: %v", err)
		}
	}

	return nil
}
