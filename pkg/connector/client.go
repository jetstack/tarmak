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
	providerSocket = "tarmak-provider.sock"
)

func (c *Connector) ConnectClient() error {

	expBackoff := backoff.NewExponentialBackOff()
	expBackoff.InitialInterval = time.Second
	expBackoff.MaxElapsedTime = time.Minute

	ctx, _ := context.WithCancel(context.Background())
	b := backoff.WithContext(expBackoff, ctx)

	resolveClient := func() error {
		client, err := rpc.Dial("unix", providerSocket)
		if err != nil {
			return err
		}

		c.client = client
		return nil
	}

	if err := backoff.Retry(resolveClient, b); err != nil {
		return fmt.Errorf("unable to resolve tarmak-provider client: %v", err)
	}

	return nil
}
