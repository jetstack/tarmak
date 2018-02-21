// Copyright Jetstack Ltd. See LICENSE for details.
package connector

import (
	"context"
	"fmt"
	"net/rpc"
	"time"

	"github.com/cenkalti/backoff"
)

const (
	tarmakSocket = "tarmak.sock"
)

func (c *Connector) ConnectorClient() error {
	expBackoff := backoff.NewExponentialBackOff()
	expBackoff.InitialInterval = time.Second
	expBackoff.MaxElapsedTime = time.Minute * 2

	b := backoff.WithContext(expBackoff, context.Background())

	resolveClient := func() error {
		client, err := rpc.Dial("unix", tarmakSocket)
		if err != nil {
			fmt.Printf("unable to connect to unix socket '%s': %v\n", tarmakSocket, err)
			return err
		}

		c.client = client
		return nil
	}

	if err := backoff.Retry(resolveClient, b); err != nil {
		return fmt.Errorf("unable to resolve tarmak RPC client: %v", err)
	}

	return nil
}

func (c *Connector) CallInit() ([]byte, error) {
	var args string
	var reply string

	if err := c.client.Call("Tarmak.Init", args, &reply); err != nil {
		return nil, fmt.Errorf("failed to call init to tarmak rpc server: %v", err)
	}

	return []byte(reply), nil
}

func (c *Connector) CloseClient() error {
	if err := c.client.Close(); err != nil {
		return fmt.Errorf("failed to close connector client: %v", err)
	}

	return nil
}
