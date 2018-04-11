// Copyright Jetstack Ltd. See LICENSE for details.
package cmd

import (
	"github.com/hashicorp/terraform/plugin"
	"github.com/spf13/cobra"

	"github.com/jetstack/tarmak/pkg/terraform/providers/tarmak"
)

var providerTarmakCmd = &cobra.Command{
	Use:   "provider-tarmak",
	Short: "launch terraform-provider-tarmak",
	Run: func(cmd *cobra.Command, args []string) {
		if err := tarmak.StartClient(); err != nil {
			panic(err)
		}

		plugin.Serve(&plugin.ServeOpts{
			ProviderFunc: tarmak.Provider,
		})
	},
}

func init() {
	subCommands = append(subCommands, providerTarmakCmd)
}
