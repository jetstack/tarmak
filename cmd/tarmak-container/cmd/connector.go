// Copyright Jetstack Ltd. See LICENSE for details.
package cmd

import (
	"github.com/spf13/cobra"

	"github.com/jetstack/tarmak/pkg/tarmak/connector"
	"github.com/jetstack/tarmak/pkg/terraform/providers/tarmak/rpc"
)

var connectorCmd = &cobra.Command{
	Use:   "connector",
	Short: "Launch tarmak connector",
	Long:  "Launch tarmak connector",
	RunE: func(cmd *cobra.Command, args []string) error {
		p := connector.NewProxy(rpc.ConnectorSocket)

		err := p.Start()
		if err != nil {
			return err
		}
		<-p.Done
		return nil
	},
}

func init() {
	subCommands = append(subCommands, connectorCmd)
}
