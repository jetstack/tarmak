package cmd

import (
	"github.com/spf13/cobra"

	"github.com/jetstack/tarmak/pkg/tarmak"
)

var environmentInitCmd = &cobra.Command{
	Use:     "init",
	Aliases: []string{"initialise", "initialize"},
	Short:   "initialize a environment",
	Run: func(cmd *cobra.Command, args []string) {
		t := tarmak.New(cmd)
		t.Must(t.CmdEnvironmentInit())
	},
}

func init() {
	environmentCmd.AddCommand(environmentInitCmd)
}
