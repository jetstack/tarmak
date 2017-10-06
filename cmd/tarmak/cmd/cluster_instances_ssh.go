package cmd

import (
	"github.com/spf13/cobra"

	"github.com/jetstack/tarmak/pkg/tarmak"
)

var clusterInstancesSshCmd = &cobra.Command{
	Use:   "ssh",
	Short: "Log into an instance with SSH",
	Run: func(cmd *cobra.Command, args []string) {
		t := tarmak.New(cmd)
		t.SSHPassThrough(args)
	},
}

func init() {
	clusterInstancesCmd.AddCommand(clusterInstancesSshCmd)
}
