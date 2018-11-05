// Copyright Jetstack Ltd. See LICENSE for details.
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/jetstack/tarmak/pkg/tarmak"
	"github.com/jetstack/tarmak/pkg/tarmak/snapshot/consul"
)

var clusterSnapshotConsulRestoreCmd = &cobra.Command{
	Use:   "restore [source path]",
	Short: "restore consul cluster with source snapshot",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("expecting single source path, got=%d", len(args))
		}

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		t := tarmak.New(globalFlags)
		c := consul.NewConsul(t, args[0])
		t.CancellationContext().WaitOrCancel(t.NewCmdSnapshot(cmd.Flags(), args, c).Restore)
	},
}

func init() {
	clusterSnapshotConsulCmd.AddCommand(clusterSnapshotConsulRestoreCmd)
}
