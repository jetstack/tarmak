// Copyright Jetstack Ltd. See LICENSE for details.
package awstag

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/cenkalti/backoff"
)

const (
	providerSocket = "tarmak-connector"
)

type ConnectorClient struct {
	client net.Conn
	reader *bufio.Reader
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
			fmt.Printf("unable to connect to uinx socket '%s': %v", providerSocket, err)
			return err
		}

		client = conn
		return nil
	}

	if err := backoff.Retry(resolveClient, b); err != nil {
		return nil, fmt.Errorf("failed to resolve connector - provider client: %v", err)
	}

	return &ConnectorClient{client, bufio.NewReader(client)}, nil
}

func (c *ConnectorClient) CloseClient() error {
	return c.client.Close()
}

func (c *ConnectorClient) ReadBytes() ([]byte, error) {
	var buff []byte

LOOP:
	for {
		b, err := c.reader.ReadByte()

		switch err {
		case io.EOF:
			break LOOP

		case nil:
			buff = append(buff, b)

		default:
			return nil, fmt.Errorf("failed to read byte from server: %v", err)
		}
	}

	return buff, nil
}
