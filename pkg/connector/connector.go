// Copyright Jetstack Ltd. See LICENSE for details.
package connector

import (
	"fmt"
	"net"
	"net/rpc"

	"github.com/hashicorp/go-multierror"
	"github.com/spf13/cobra"
)

type Connector struct {
	client *rpc.Client
	server net.Listener
}

func (c *Connector) InitiateConnection() error {
	if err := c.NewClient(); err != nil {
		return err
	}

	if err := c.NewServer(); err != nil {
		return err
	}

	reply, err := c.CallInit()
	if err != nil {
		return err
	}

	conn, err := c.AcceptProvider()
	if err != nil {
		return err
	}

	return c.SendProvider(conn, reply)
}

func NewCommandStartConnector(stopCh chan struct{}) *cobra.Command {
	cmd := &cobra.Command{
		Short: "Launch tarmak connector",
		Long:  "Launch tarmak connector",
		RunE: func(c *cobra.Command, args []string) error {
			var result *multierror.Error
			connector := new(Connector)

			if err := connector.InitiateConnection(); err != nil {
				return fmt.Errorf("error initialising connection: %v", err)
			}

			go func() {
				if err := connector.StartServer(stopCh); err != nil {
					fmt.Printf("error in connector server: %v", err)
				}
			}()

			<-stopCh

			if err := connector.CloseClient(); err != nil {
				result = multierror.Append(result, err)
			}

			if err := connector.CloseServer(); err != nil {
				result = multierror.Append(result, err)
			}

			return result.ErrorOrNil()
		},
	}

	return cmd
}
