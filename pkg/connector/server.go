// Copyright Jetstack Ltd. See LICENSE for details.
package connector

import (
	"bufio"
	"fmt"
	"io"
	"net"
)

const (
	providerSocket = "provider-tarmak.sock"
)

func (c *Connector) NewServer() error {
	ln, err := net.Listen("unix", providerSocket)
	if err != nil {
		return fmt.Errorf("unable to listen to provider socket: %v", err)
	}

	c.server = ln

	return nil
}

func (c *Connector) CloseServer() error {
	if err := c.server.Close(); err != nil {
		return fmt.Errorf("failed to close connector server: %v", err)
	}

	return nil
}

func (c *Connector) StartServer(stopCh chan struct{}) error {
	for {
		select {
		case <-stopCh:
			return nil
		default:
			conn, err := c.AcceptProvider()
			if err != nil {
				return err
			}

			bytes, err := c.HandleConnection(conn)
			if err != nil {
				return err
			}

			fmt.Printf("Received from client: %s\n", bytes)
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

	reader := bufio.NewReader(conn)

LOOP:
	for {
		b, err := reader.ReadByte()

		switch err {
		case io.EOF:
			break LOOP

		case nil:
			buff = append(buff, b)

		default:
			return nil, fmt.Errorf("failed to read byte from client: %v", err)
		}
	}

	return buff, nil
}

func (c *Connector) SendProvider(conn net.Conn, bytes []byte) error {
	writer := bufio.NewWriter(conn)

	for _, b := range bytes {
		if err := writer.WriteByte(b); err != nil {
			return fmt.Errorf("error sending bytes to connection: %v", err)
		}
	}

	return nil
}
