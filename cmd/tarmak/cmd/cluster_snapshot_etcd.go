// Copyright Jetstack Ltd. See LICENSE for details.
package cmd

import (
	"github.com/spf13/cobra"
)

var clusterSnapshotEtcdCmd = &cobra.Command{
	Use:   "etcd",
	Short: "Manage snapshots on remote etcd clusters",
}

func init() {
	clusterSnapshotCmd.AddCommand(clusterSnapshotEtcdCmd)
}
