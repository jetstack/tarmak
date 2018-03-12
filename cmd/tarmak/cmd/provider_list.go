// Copyright Jetstack Ltd. See LICENSE for details.
package cmd

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/jetstack/tarmak/pkg/tarmak"
	"github.com/jetstack/tarmak/pkg/tarmak/utils"
)

var providerListCmd = &cobra.Command{
	Use:   "list",
	Short: "Print a list of providers",
	Run: func(cmd *cobra.Command, args []string) {
		t := tarmak.New(globalFlags)
		defer t.Cleanup()
		varMaps := make([]map[string]string, 0)
		for _, prov := range t.Providers() {
			varMaps = append(varMaps, prov.Parameters())
		}
		utils.ListParameters(os.Stdout, []string{"name", "cloud"}, varMaps)
	},
}

func init() {
	providerCmd.AddCommand(providerListCmd)
}
