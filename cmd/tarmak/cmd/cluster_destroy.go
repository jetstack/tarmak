package cmd

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/jetstack/tarmak/pkg/tarmak"
	"github.com/jetstack/tarmak/pkg/tarmak/utils"
)

// clusterDestroyCmd handles `tarmak clusters destroy`
var clusterDestroyCmd = &cobra.Command{
	Use:   "destroy",
	Short: "Destroy the current cluster",
	Run: func(cmd *cobra.Command, args []string) {
		t := tarmak.New(cmd)
		utils.WaitOrCancel(
			func(ctx context.Context) error {
				return t.CmdTerraformDestroy(args, ctx)
			},
		)
	},
}

func init() {
	terraformPFlags(clusterDestroyCmd.PersistentFlags())
	clusterDestroyCmd.PersistentFlags().Bool(tarmak.FlagForceDestroyStateStack, false, "destroy the state stack as well, possibly dangerous")
	clusterCmd.AddCommand(clusterDestroyCmd)
}
