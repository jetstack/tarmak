// Copyright Jetstack Ltd. See LICENSE for details.
package cmd

import (
	"github.com/spf13/cobra"
)

var clusterSnapshotEtcdSaveCmd = &cobra.Command{
	Use:   "save [target path]",
	Short: "save etcd snapshot to target path",
	Run: func(cmd *cobra.Command, args []string) {
	},
}

func init() {
	clusterSnapshotEtcdCmd.AddCommand(clusterSnapshotEtcdSaveCmd)
}
