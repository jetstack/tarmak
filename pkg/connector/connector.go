// Copyright Jetstack Ltd. See LICENSE for details.
package connector

import (
	"net"
	"net/rpc"

	"github.com/hashicorp/go-multierror"
	"github.com/spf13/cobra"
)

type Connector struct {
	client *rpc.Client
	server net.Conn
}

func (c *Connector) InitiateConnection() error {
	reply, err := c.CallInit()
	if err != nil {
		return err
	}

	return c.ForwardConnection(reply)
}

func NewCommandStartConnector(stopCh chan struct{}) *cobra.Command {
	cmd := &cobra.Command{
		Short: "Launch tarmak connector",
		Long:  "Launch tarmak connector",
		RunE: func(c *cobra.Command, args []string) error {
			var result *multierror.Error
			connector := new(Connector)

			if err := connector.ConnectorClient(); err != nil {
				result = multierror.Append(result, err)
			}

			if err := connector.StartServer(); err != nil {
				result = multierror.Append(result, err)
			}

			if result != nil {
				return result
			}

			if err := connector.InitiateConnection(); err != nil {
				result = multierror.Append(result, err)
			}

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
