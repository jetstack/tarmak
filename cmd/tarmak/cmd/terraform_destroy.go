package cmd

import (
	"github.com/spf13/cobra"

	"github.com/jetstack/tarmak/pkg/tarmak"
)

// tfdestroyCmd represents the tfdestroy command
var terraformDestroyCmd = &cobra.Command{
	Use:     "terraform-destroy",
	Aliases: []string{"t-d"},
	Short:   "This applies the set of stacks in the current cluster",
	Run: func(cmd *cobra.Command, args []string) {
		t := tarmak.New(cmd)
		t.Must(t.CmdTerraformDestroy(args))
	},
}

func init() {
	terraformPFlags(terraformDestroyCmd.PersistentFlags())
	terraformDestroyCmd.PersistentFlags().Bool(tarmak.FlagForceDestroyStateStack, false, "destroy the state stack as well, possibly dangerous")
	RootCmd.AddCommand(terraformDestroyCmd)
}
