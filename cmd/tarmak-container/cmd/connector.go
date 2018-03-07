// Copyright Jetstack Ltd. See LICENSE for details.
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/jetstack/tarmak/pkg/connector"
)

var connectorCmd = &cobra.Command{
	Use:   "connector",
	Short: "Launch tarmak connector",
	Long:  "Launch tarmak connector",
	RunE: func(cmd *cobra.Command, args []string) error {
		connector := connector.NewConnector(newLogger())

		if err := connector.StartConnector(); err != nil {
			return fmt.Errorf("connector failed: %v", err)
		}

		return nil
	},
}

func init() {
	subCommands = append(subCommands, connectorCmd)
}
