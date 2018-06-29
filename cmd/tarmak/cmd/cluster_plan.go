// Copyright Jetstack Ltd. See LICENSE for details.
package cmd

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/jetstack/tarmak/pkg/tarmak"
	"github.com/jetstack/tarmak/pkg/tarmak/utils"
)

var clusterPlanCmd = &cobra.Command{
	Use:   "plan",
	Short: "Plan changes on the currently configured cluster",
	Run: func(cmd *cobra.Command, args []string) {
		t := tarmak.New(globalFlags)
		defer t.Cleanup()

		utils.WaitOrCancel(
			func(ctx context.Context) error {
				return t.NewCmdTerraform(args).Plan()
			},
			t.Context(),
			2,
		)
	},
}

func init() {
	clusterCmd.AddCommand(clusterPlanCmd)
}
