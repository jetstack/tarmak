// Copyright Jetstack Ltd. See LICENSE for details.
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/jetstack/tarmak/pkg/tarmak"
	"github.com/jetstack/tarmak/pkg/tarmak/snapshot/etcd"
	"github.com/jetstack/tarmak/pkg/tarmak/utils/consts"
)

var clusterSnapshotEtcdRestoreCmd = &cobra.Command{
	Use:   "restore",
	Short: "restore etcd cluster with source snapshots",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if !cmd.Flags().Changed(consts.RestoreK8sMainFlagName) &&
			!cmd.Flags().Changed(consts.RestoreK8sEventsFlagName) &&
			!cmd.Flags().Changed(consts.RestoreOverlayFlagName) {

			return fmt.Errorf("expecting at least one set flag of [%s %s %s]",
				consts.RestoreK8sMainFlagName,
				consts.RestoreK8sEventsFlagName,
				consts.RestoreOverlayFlagName,
			)
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		t := tarmak.New(globalFlags)
		s := etcd.New(t, "")
		t.CancellationContext().WaitOrCancel(t.NewCmdSnapshot(cmd.Flags(), args, s).Restore)
	},
}

func init() {
	clusterSnapshotEtcdRestoreFlags(
		clusterSnapshotEtcdRestoreCmd.PersistentFlags(),
	)
	clusterSnapshotEtcdCmd.AddCommand(clusterSnapshotEtcdRestoreCmd)
}
