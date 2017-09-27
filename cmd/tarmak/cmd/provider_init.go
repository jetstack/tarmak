package cmd

import (
	"github.com/spf13/cobra"

	"github.com/jetstack/tarmak/pkg/tarmak"
)

var providerInitCmd = &cobra.Command{
	Use:     "init",
	Aliases: []string{"initialise", "initialize"},
	Short:   "initialize a provider",
	Run: func(cmd *cobra.Command, args []string) {
		t := tarmak.New(cmd)
		t.Must(t.CmdProviderInit())
	},
}

func init() {
	providerCmd.AddCommand(providerInitCmd)
}
