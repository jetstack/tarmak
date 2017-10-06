// Copyright Jetstack Ltd. See LICENSE for details.
package cmd

import (
	"github.com/spf13/cobra"

	"github.com/jetstack/tarmak/pkg/tarmak"
)

var clusterSshCmd = &cobra.Command{
	Use:   "ssh",
	Short: "Log into an instance with SSH",
	Run: func(cmd *cobra.Command, args []string) {
		t := tarmak.New(globalFlags)
		t.SSHPassThrough(args)
	},
}

func init() {
	clusterCmd.AddCommand(clusterSshCmd)
}
