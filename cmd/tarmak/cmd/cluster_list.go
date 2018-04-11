// Copyright Jetstack Ltd. See LICENSE for details.
package cmd

import (
	"os"

	"github.com/spf13/cobra"

	clusterv1alpha1 "github.com/jetstack/tarmak/pkg/apis/cluster/v1alpha1"
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
				kubernetesVersion := ""
				if cluster.Type() != clusterv1alpha1.ClusterTypeHub {
					kubernetesVersion = cluster.Config().Kubernetes.Version
				}

				current := "false"
				if t.Cluster().Name() == cluster.Name() && t.Cluster().Environment().Name() == t.Cluster().Environment().Name() {
					current = "true"
				}

				varMaps = append(varMaps, map[string]string{
					"name":        cluster.Name(),
					"environment": cluster.Environment().Name(),
					"version":     kubernetesVersion,
					"type":        cluster.Type(),
					"zone":        cluster.Variables()["public_zone"].(string),
					"current":     current,
				})
			}
		}
		utils.ListParameters(os.Stdout, []string{"name", "environment", "zone", "type", "version", "current"}, varMaps)
	},
}

func init() {
	clusterCmd.AddCommand(clusterListCmd)
}
