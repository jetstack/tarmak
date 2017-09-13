package cmd

import (
	"github.com/spf13/cobra"

	"github.com/jetstack/tarmak/pkg/tarmak"
)

var clusterApplyCmd = &cobra.Command{
	Use:     "apply",
	Aliases: []string{"t-a"},
	Short:   "This applies the set of stacks in the current context",
	Run: func(cmd *cobra.Command, args []string) {
		t := tarmak.New(cmd)
		t.Must(t.CmdTerraformApply(args))
	},
}

func init() {
	terraformPFlags(clusterApplyCmd.PersistentFlags())
	clusterCmd.AddCommand(clusterApplyCmd)
}
