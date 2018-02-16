// Copyright Jetstack Ltd. See LICENSE for details.
package connector

import (
	"net/rpc"

	"github.com/spf13/cobra"
)

type Connector struct {
	client  *rpc.Client
	server  *rpc.Server
	connRPC *ConnectorRPC
}

func NewCommandStartConnector(stopCh chan struct{}) *cobra.Command {
	cmd := &cobra.Command{
		Short: "Launch tarmak connector",
		Long:  "Launch tarmak connector",
		RunE: func(c *cobra.Command, args []string) error {

			connector := &Connector{}
			if err := connector.ConnectClient(); err != nil {
				return err
			}

			go func() {
				connector.StartServer()
			}()

			<-stopCh

			return nil
		},
	}

	return cmd
}
