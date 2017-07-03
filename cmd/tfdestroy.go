package cmd

import (
	"github.com/spf13/cobra"

	"github.com/jetstack/tarmak/pkg/tarmak"
)

// tfdestroyCmd represents the tfdestroy command
var tfdestroyCmd = &cobra.Command{
	Use:   "tfdestroy",
	Short: "This applies the set of stacks in the current context",
	Run: func(cmd *cobra.Command, args []string) {
		t := tarmak.New()
		t.TerraformDestroy()
	},
}

func init() {
	RootCmd.AddCommand(tfdestroyCmd)
}
