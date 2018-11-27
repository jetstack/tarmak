// Copyright Jetstack Ltd. See LICENSE for details.
package cmd

import (
	"github.com/spf13/cobra"
)

var clusterSnapshotCmd = &cobra.Command{
	Use:   "snapshot",
	Short: "Manage snapshots of remote consul and etcd clusters",
}

func init() {
	clusterCmd.AddCommand(clusterSnapshotCmd)
}
