// Copyright Jetstack Ltd. See LICENSE for details.
package cmd

import (
	"github.com/spf13/cobra"

	"github.com/jetstack/tarmak/pkg/tarmak"
)

var clusterInstancesSshCmd = &cobra.Command{
	Use:   "ssh [instance alias] [optional ssh command]",
	Short: "Log into an instance with SSH",
	Run: func(cmd *cobra.Command, args []string) {
		t := tarmak.New(globalFlags)
		defer t.Cleanup()
		t.SSHPassThrough(args)
	},
	DisableFlagsInUseLine: true,
}

func init() {
	clusterInstancesCmd.AddCommand(clusterInstancesSshCmd)
}
