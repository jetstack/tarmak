// Copyright Jetstack Ltd. See LICENSE for details.
package cmd

import (
	"github.com/spf13/cobra"

	"github.com/jetstack/tarmak/pkg/tarmak"
	"github.com/jetstack/tarmak/pkg/tarmak/snapshot/etcd"
)

var clusterSnapshotEtcdSaveCmd = &cobra.Command{
	Use:   "save [target path prefix]",
	Short: "save etcd snapshot to target path prefix, i.e 'backup-'",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			args = []string{""}
		}
		t := tarmak.New(globalFlags)
		s := etcd.New(t, args[0])
		t.CancellationContext().WaitOrCancel(t.NewCmdSnapshot(cmd.Flags(), args, s).Save)
	},
}

func init() {
	clusterSnapshotEtcdCmd.AddCommand(clusterSnapshotEtcdSaveCmd)
}
