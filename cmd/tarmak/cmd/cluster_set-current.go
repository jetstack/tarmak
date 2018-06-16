// Copyright Jetstack Ltd. See LICENSE for details.
package cmd

import (
	"github.com/spf13/cobra"

	"github.com/jetstack/tarmak/pkg/tarmak"
)

var clusterSetCurrentCmd = &cobra.Command{
	Use:   "set-current",
	Short: "Set current cluster in config",
	Run: func(cmd *cobra.Command, args []string) {
		t := tarmak.New(globalFlags)
		defer t.Cleanup()

		if len(args) != 1 {
			t.Log().Fatalf("Only one cluster can be set as current cluster")
		}

		found := false
	LOOP:
		for _, env := range t.Environments() {
			for _, cluster := range env.Clusters() {
				if args[0] == cluster.ClusterName() {
					found = true
					break LOOP
				}
			}
		}

		if !found {
			t.Log().Fatalf("Failed to find cluster '%s' in config", args[0])
		}

		if err := t.Config().SetCurrentCluster(args[0]); err != nil {
			t.Log().Fatalf("Failed to set current cluster in config: %v", err)
		}

	},
}

func init() {
	clusterCmd.AddCommand(clusterSetCurrentCmd)
}
