// Copyright Jetstack Ltd. See LICENSE for details.
package cmd

import (
	"github.com/spf13/cobra"
)

var clusterDebugCmd = &cobra.Command{
	Use:   "debug",
	Short: "Operations for debugging a cluster",
}

func init() {
	clusterCmd.AddCommand(clusterDebugCmd)
}
