// Copyright Jetstack Ltd. See LICENSE for details.
package awstag

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"time"

	"github.com/cenkalti/backoff"
)

const (
	providerSocket = "provider.sock"
	EOT            = byte(4)
)

type ConnectorClient struct {
	client net.Conn
}

func StartClient() error {
	client, err := NewClient()
	if err != nil {
		return err
	}

	bytes, err := client.ReadBytes()
	if err != nil {
		return err
	}

	fmt.Printf("Received bytes from connector: %s\n", bytes)

	return nil
}

func NewClient() (*ConnectorClient, error) {
	var client net.Conn

	expBackoff := backoff.NewExponentialBackOff()
	expBackoff.InitialInterval = time.Second
	expBackoff.MaxElapsedTime = time.Minute * 2

	b := backoff.WithContext(expBackoff, context.Background())

	resolveClient := func() error {
		conn, err := net.Dial("unix", providerSocket)
		if err != nil {
			fmt.Printf("unable to dial into uinx socket '%s': %v\n", providerSocket, err)
			return err
		}

		client = conn
		return nil
	}

	if err := backoff.Retry(resolveClient, b); err != nil {
		return nil, fmt.Errorf("failed to resolve provider client: %v", err)
	}

	return &ConnectorClient{client}, nil
}

func (c *ConnectorClient) CloseClient() error {
	return c.client.Close()
}

func (c *ConnectorClient) ReadBytes() ([]byte, error) {
	var buff []byte
	b := make([]byte, 1)

LOOP:
	for {
		_, err := c.client.Read(b)
		if err != nil {
			return nil, fmt.Errorf("failed to read byte from server: %v", err)
		}

		if b[0] == EOT {
			break LOOP
		}

		buff = append(buff, b...)
	}

	return buff, nil
}

func (c *ConnectorClient) SendBytes(bytes []byte) error {
	writer := bufio.NewWriter(c.client)

	for _, b := range bytes {
		if err := writer.WriteByte(b); err != nil {
			return fmt.Errorf("error sending bytes to connector server: %v", err)
		}
	}

	return nil
}
