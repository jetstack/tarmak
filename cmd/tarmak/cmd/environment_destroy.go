// Copyright Jetstack Ltd. See LICENSE for details.
package cmd

import (
	"errors"

	"github.com/jetstack/tarmak/pkg/tarmak"
	"github.com/spf13/cobra"
)

// environmentDestroyCmd handles `tarmak environment destroy`
var environmentDestroyCmd = &cobra.Command{
	Use:   "destroy [name]",
	Short: "Destroy an environment",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return errors.New("you have to give one environment name")
		}
		return nil
	},

	Run: func(cmd *cobra.Command, args []string) {
		t := tarmak.New(globalFlags)

		t.CancellationContext().WaitOrCancel(t.NewCmdTarmak(cmd.Flags(), args).DestroyEnvironment)
	},
}

func init() {
	environmentDestroyFlags(environmentDestroyCmd.PersistentFlags())
	environmentCmd.AddCommand(environmentDestroyCmd)
}
