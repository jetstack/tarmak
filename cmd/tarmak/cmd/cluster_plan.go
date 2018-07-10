// Copyright Jetstack Ltd. See LICENSE for details.
package cmd

import (
	"github.com/spf13/cobra"

	"github.com/jetstack/tarmak/pkg/tarmak"
)

var clusterPlanCmd = &cobra.Command{
	Use:   "plan",
	Short: "Plan changes on the currently configured cluster",
	Run: func(cmd *cobra.Command, args []string) {
		t := tarmak.New(globalFlags)
		defer t.Cleanup()

		t.Context().WaitOrCancel(
			t.NewCmdTerraform(args).Plan,
			2,
		)
	},
}

func init() {
	clusterCmd.AddCommand(clusterPlanCmd)
}
