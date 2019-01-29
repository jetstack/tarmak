// Copyright Jetstack Ltd. See LICENSE for details.
package cmd

import (
	"errors"

	"github.com/spf13/cobra"

	"github.com/jetstack/tarmak/pkg/tarmak"
)

var clusterSshCmd = &cobra.Command{
	Use:   "ssh [instance alias] [optional ssh arguments]",
	Short: "Log into an instance with SSH",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("expecting an instance aliases argument")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		t := tarmak.New(globalFlags)
		t.Perform(t.SSHPassThrough(args))
	},
	DisableFlagsInUseLine: true,
}

func init() {
	clusterCmd.AddCommand(clusterSshCmd)
}
