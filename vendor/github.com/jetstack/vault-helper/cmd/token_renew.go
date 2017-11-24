package cmd

import (
	"github.com/spf13/cobra"
)

// initCmd represents the init command
var renewtokenCmd = &cobra.Command{
	Use:   "renew-token",
	Short: "Renew token on vault server.",
	Run: func(cmd *cobra.Command, args []string) {

		i, err := newInstanceToken(cmd)
		if err != nil {
			i.Log.Fatal(err)
		}

		if err := i.TokenRenewRun(); err != nil {
			i.Log.Fatal(err)
		}
	},
}

func init() {
	instanceTokenFlags(renewtokenCmd)
	RootCmd.AddCommand(renewtokenCmd)
}
