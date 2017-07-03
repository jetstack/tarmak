package cmd

import (
	"github.com/spf13/cobra"

	"github.com/jetstack/tarmak/pkg/tarmak"
)

// tfapplyCmd represents the tfapply command
var tfapplyCmd = &cobra.Command{
	Use:   "tfapply",
	Short: "This applies the set of stacks in the current context",
	Run: func(cmd *cobra.Command, args []string) {
		t := tarmak.New()
		t.TerraformApply()
	},
}

func init() {
	RootCmd.AddCommand(tfapplyCmd)
}
