package cmd

import (
	"github.com/spf13/cobra"

	"github.com/jetstack/tarmak/pkg/tarmak"
)

var terraformApplyCmd = &cobra.Command{
	Use:     "terraform-apply",
	Aliases: []string{"t-a"},
	Short:   "This applies the set of stacks in the current context",
	Run: func(cmd *cobra.Command, args []string) {
		t := tarmak.New(cmd)
		t.TerraformApply(args)
	},
}

func init() {
	terraformPFlags(terraformApplyCmd.PersistentFlags())
	RootCmd.AddCommand(terraformApplyCmd)
}
