// Copyright Jetstack Ltd. See LICENSE for details.
package cmd

import (
	"github.com/hashicorp/terraform/plugin"
	"github.com/spf13/cobra"

	"github.com/jetstack/tarmak/pkg/terraform/providers/awstag"
)

var providerAWSTagCmd = &cobra.Command{
	Use:   "provider-awstag",
	Short: "launch terraform-provider-awstag",
	Run: func(cmd *cobra.Command, args []string) {
		plugin.Serve(&plugin.ServeOpts{
			ProviderFunc: awstag.Provider,
		})
	},
}

func init() {
	subCommands = append(subCommands, providerAWSTagCmd)
}
