// Copyright Jetstack Ltd. See LICENSE for details.
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/jetstack/tarmak/pkg/tarmak"
)

var clusterSetCurrentCmd = &cobra.Command{
	Use:   "set-current [environment-cluster]",
	Short: "Set current cluster in config",
	Run: func(cmd *cobra.Command, args []string) {
		t := tarmak.New(globalFlags)

		if len(args) != 1 {
			t.Log().Fatal("Expecting a single environment-cluster argument to be set as current")
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
			t.Perform(fmt.Errorf("Failed to find cluster '%s' in config", args[0]))
		}

		t.Perform(t.Config().SetCurrentCluster(args[0]))
	},
}

func init() {
	clusterCmd.AddCommand(clusterSetCurrentCmd)
}
