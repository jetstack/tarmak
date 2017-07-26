package cmd

import (
	"github.com/spf13/cobra"

	"github.com/jetstack/tarmak/pkg/tarmak"
)

var kubectlCmd = &cobra.Command{
	Use:   "kubectl",
	Short: "kubectl against the current cluster",
	Run: func(cmd *cobra.Command, args []string) {
		t := tarmak.New(cmd)
		t.Must(t.CmdKubectl(args))
	},
	DisableFlagParsing: true,
}

func init() {
	RootCmd.AddCommand(kubectlCmd)
}
