// Copyright Jetstack Ltd. See LICENSE for details.
package connector

import (
	"fmt"
	"net/rpc"
)

const (
	providerSocket = "tarmak-provider.sock"
)

func (c *Connector) ConnectClient() error {
	client, err := rpc.Dial("unix", providerSocket)
	if err != nil {
		return fmt.Errorf("failed to resolve provide socket: %v", err)
	}

	c.client = client

	return nil
}
