package cmd

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/jetstack/tarmak/pkg/tarmak"
	"github.com/jetstack/tarmak/pkg/tarmak/utils"
)

var clusterApplyCmd = &cobra.Command{
	Use:     "apply",
	Aliases: []string{"t-a"},
	Short:   "This applies the set of stacks in the current cluster",
	Run: func(cmd *cobra.Command, args []string) {
		t := tarmak.New(cmd)
		utils.WaitOrCancel(
			func(ctx context.Context) error {
				return t.CmdTerraformApply(args, ctx)
			},
		)
	},
}

func init() {
	terraformPFlags(clusterApplyCmd.PersistentFlags())
	clusterCmd.AddCommand(clusterApplyCmd)
}
