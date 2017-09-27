package cmd

import (
	"github.com/spf13/cobra"

	"github.com/jetstack/tarmak/pkg/tarmak"
)

var clusterInitCmd = &cobra.Command{
	Use:     "init",
	Aliases: []string{"initialise", "initialize"},
	Short:   "initialize a cluster",
	Run: func(cmd *cobra.Command, args []string) {
		t := tarmak.New(cmd)
		t.Must(t.CmdClusterInit())
	},
}

func init() {
	clusterCmd.AddCommand(clusterInitCmd)
}
