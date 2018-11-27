// Copyright Jetstack Ltd. See LICENSE for details.
package cmd

import (
	"github.com/spf13/cobra"
)

var clusterSnapshotConsulCmd = &cobra.Command{
	Use:   "consul",
	Short: "Manage snapshots on remote consul clusters",
}

func init() {
	clusterSnapshotCmd.AddCommand(clusterSnapshotConsulCmd)
}
