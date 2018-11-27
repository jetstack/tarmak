// Copyright Jetstack Ltd. See LICENSE for details.
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/jetstack/tarmak/pkg/tarmak"
	"github.com/jetstack/tarmak/pkg/tarmak/snapshot/etcd"
)

var clusterSnapshotEtcdRestoreCmd = &cobra.Command{
	Use:   "restore [source path]",
	Short: "restore etcd cluster with source snapshot",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("expecting single target path, got=%d", len(args))
		}

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		t := tarmak.New(globalFlags)
		s := etcd.New(t, args[0])
		t.CancellationContext().WaitOrCancel(t.NewCmdSnapshot(cmd.Flags(), args, s).Restore)
	},
}

func init() {
	clusterSnapshotEtcdCmd.AddCommand(clusterSnapshotEtcdRestoreCmd)
}
