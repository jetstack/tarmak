package cmd

import (
	"github.com/spf13/cobra"
	//"github.com/jetstack/tarmak/pkg/tarmak"
)

var enviromentsCmd = &cobra.Command{
	Use:   "environments",
	Short: "environments resource [list | init]",
	Run: func(cmd *cobra.Command, args []string) {
		if args[0] == "list" {
			// List enviroments
		} else if args[0] == "init" {
			// init enviroment
		} else {
			// Command not recognised
		}
	},
	DisableFlagParsing: true,
}

func init() {
	RootCmd.AddCommand(enviromentsCmd)
}
