// Copyright Jetstack Ltd. See LICENSE for details.
package connector

import (
	"fmt"
	"io"
	"net"
)

const (
	providerSocket = "provider.sock"
	EOT            = byte(4)
)

func (c *Connector) NewServer() error {
	ln, err := net.Listen("unix", providerSocket)
	if err != nil {
		return fmt.Errorf("unable to listen to provider socket: %v", err)
	}

	c.server = ln

	c.log.Infof("Connector server started.")

	return nil
}

func (c *Connector) CloseServer() error {
	if err := c.server.Close(); err != nil {
		return fmt.Errorf("failed to close connector server: %v", err)
	}

	c.log.Infof("Connector server closed.")

	return nil
}

func (c *Connector) StartServer() {
	for {
		select {
		case <-c.stopCh:
			return
		default:
			conn, err := c.AcceptProvider()
			if err != nil {
				continue
			}

			bytes, err := c.HandleConnection(conn)
			if err != nil {
				c.log.Errorf("error handling connection: %v", err)
				continue
			}

			c.log.Debugf("Received from client: %s\n", bytes)
		}
	}
}

func (c *Connector) AcceptProvider() (net.Conn, error) {
	conn, err := c.server.Accept()
	if err != nil {
		return nil, fmt.Errorf("unable to accept from provider socket: %v", err)
	}

	return conn, nil
}

func (c *Connector) HandleConnection(conn net.Conn) ([]byte, error) {
	var buff []byte
	b := make([]byte, 1)

LOOP:
	for {
		_, err := conn.Read(b)

		switch err {
		case io.EOF:
			break LOOP

		case nil:
			buff = append(buff, b...)

		default:
			return nil, fmt.Errorf("failed to read byte from client: %v", err)
		}
	}

	return buff, nil
}

func (c *Connector) SendProvider(conn net.Conn, bytes []byte) error {
	for _, b := range bytes {
		if _, err := conn.Write([]byte{b}); err != nil {
			return fmt.Errorf("error sending byte to connection: %v", err)
		}
	}

	if _, err := conn.Write([]byte{EOT}); err != nil {
		return fmt.Errorf("error sending EOT to connection: %v", err)
	}

	c.log.Debugf("Sent bytes to provider.")

	return nil
}
