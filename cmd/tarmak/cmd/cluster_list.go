// Copyright Jetstack Ltd. See LICENSE for details.
package cmd

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/jetstack/tarmak/pkg/tarmak"
	"github.com/jetstack/tarmak/pkg/tarmak/utils"
)

var clusterListCmd = &cobra.Command{
	Use:   "list",
	Short: "Print a list of clusters",
	Run: func(cmd *cobra.Command, args []string) {
		t := tarmak.New(globalFlags)
		defer t.Cleanup()

		varMaps := make([]map[string]string, 0)
		for _, env := range t.Environments() {
			for _, cluster := range env.Clusters() {
				varMaps = append(varMaps, map[string]string{
					"name":        cluster.Name(),
					"environment": cluster.Environment().Name(),
					"version":     cluster.Config().Kubernetes.Version,
					"type":        cluster.Type(),
					"zone":        cluster.Variables()["public_zone"].(string),
				})
			}
		}
		utils.ListParameters(os.Stdout, []string{"name", "environment", "zone", "type", "version"}, varMaps)
	},
}

func init() {
	clusterCmd.AddCommand(clusterListCmd)
}
