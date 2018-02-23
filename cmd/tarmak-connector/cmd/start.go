// Copyright Jetstack Ltd. See LICENSE for details.
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/jetstack/tarmak/pkg/connector"
)

func NewCommandStartConnector() *cobra.Command {
	cmd := &cobra.Command{
		Short: "Launch tarmak connector",
		Long:  "Launch tarmak connector",
		RunE: func(cmd *cobra.Command, args []string) error {
			log := LogLevel(cmd)
			connector := connector.NewConnector(log)

			if err := connector.RunConnector(); err != nil {
				return fmt.Errorf("connector failed: %v", err)
			}

			return nil
		},
	}

	return cmd
}
