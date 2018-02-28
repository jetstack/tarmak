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

func (c *Connector) NewClient() error {
	expBackoff := backoff.NewExponentialBackOff()
	expBackoff.InitialInterval = time.Second
	expBackoff.MaxElapsedTime = time.Minute * 2

	b := backoff.WithContext(expBackoff, context.Background())

	resolveClient := func() error {
		client, err := rpc.Dial("unix", tarmakSocket)
		if err != nil {
			c.log.Debugf("unable to dial into unix socket '%s': %v", tarmakSocket, err)
			return err
		}

		c.client = client
		return nil
	}

	if err := backoff.Retry(resolveClient, b); err != nil {
		return fmt.Errorf("unable to resolve tarmak RPC client: %v", err)
	}

	c.log.Infof("Connector client resolved.")

	return nil
}

func (c *Connector) CallHandshake() (reply string, err error) {
	var args string

	if err := c.client.Call("Tarmak.Handshake", args, &reply); err != nil {
		return "", fmt.Errorf("failed to call handshake to tarmak rpc server: %v", err)
	}

	return reply, nil
}

func (c *Connector) CallBastionInstanceStatus(args []string) (reply string, err error) {
	if len(args) != 2 {
		return "", fmt.Errorf("CallBastionInstanceStatus expects 2 arguments. got=%d", len(args))
	}

	if err := c.client.Call("Tarmak.BastionInstanceStatus", args, &reply); err != nil {
		return "", fmt.Errorf("failed to call BastionInstanceStatus to tarmak rpc server: %v", err)
	}

	return reply, nil
}

func (c *Connector) CallVaultClusterStatus(args []string) (reply string, err error) {
	if err := c.client.Call("Tarmak.VaultClusterStatus", args, &reply); err != nil {
		return "", fmt.Errorf("failed to call VaultClusterStatus to tarmak rpc server: %v", err)
	}

	return reply, nil
}

func (c *Connector) CallVaultInstanceRoleStatus(args []string) (reply string, err error) {
	if len(args) != 1 {
		return "", fmt.Errorf("VaultInstanceRoleStatus expects 1 argument. got=%d", len(args))
	}

	if err := c.client.Call("Tarmak.VaultInstanceRoleStatus", args, &reply); err != nil {
		return "", fmt.Errorf("failed to call VaultInstanceRoleStatus to tarmak rpc server: %v", err)
	}

	return reply, nil
}

func (c *Connector) CloseClient() error {
	if err := c.client.Close(); err != nil {
		return fmt.Errorf("failed to close connector client: %v", err)
	}

	c.log.Infof("Connector client closed.")

	return nil
}
