// Copyright Jetstack Ltd. See LICENSE for details.
package awstag

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/cenkalti/backoff"
)

const (
	providerSocket = "tarmak-connector"
)

func newClient() (net.Conn, error) {
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

	return client, nil
}
