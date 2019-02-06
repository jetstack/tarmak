// Copyright Jetstack Ltd. See LICENSE for details.
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/jetstack/tarmak/pkg/tarmak"
)

var clusterForceUnlockCmd = &cobra.Command{
	Use:   "force-unlock [lock ID]",
	Short: "Remove remote lock using lock ID",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("expected single lock ID argument, got=%d", len(args))
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		t := tarmak.New(globalFlags)
		t.CancellationContext().WaitOrCancel(t.NewCmdTarmak(cmd.Flags(), args).ForceUnlock)
	},
}

func init() {
	clusterCmd.AddCommand(clusterForceUnlockCmd)
}
