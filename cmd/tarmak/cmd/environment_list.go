package cmd

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/jetstack/tarmak/pkg/tarmak"
	"github.com/jetstack/tarmak/pkg/tarmak/utils"
)

var environmentListCmd = &cobra.Command{
	Use:   "list",
	Short: "Print a list of environments",
	Run: func(cmd *cobra.Command, args []string) {
		t := tarmak.New(cmd)
		varMaps := make([]map[string]string, 0)
		for _, env := range t.Environments() {
			varMaps = append(varMaps, env.Parameters())
		}
		utils.ListParameters(os.Stdout, []string{"name", "provider", "location"}, varMaps)
	},
}

func init() {
	environmentCmd.AddCommand(environmentListCmd)
}
