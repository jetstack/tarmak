package cmd

import (
	"github.com/spf13/cobra"

	"github.com/jetstack/tarmak/pkg/tarmak"
)

var sshCmd = &cobra.Command{
	Use:   "ssh",
	Short: "ssh into instance",
	Run: func(cmd *cobra.Command, args []string) {
		t := tarmak.New(cmd)
		t.SSH(args)
	},
}

func init() {
	RootCmd.AddCommand(sshCmd)
}
