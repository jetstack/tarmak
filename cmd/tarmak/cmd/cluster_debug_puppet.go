// Copyright Jetstack Ltd. See LICENSE for details.
package cmd

import (
	"github.com/spf13/cobra"
)

var clusterDebugPuppetCmd = &cobra.Command{
	Use:   "puppet",
	Short: "Operations for debugging Puppet configuration",
}

func init() {
	clusterDebugCmd.AddCommand(clusterDebugPuppetCmd)
}
