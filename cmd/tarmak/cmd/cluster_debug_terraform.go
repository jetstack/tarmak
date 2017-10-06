// Copyright Jetstack Ltd. See LICENSE for details.
package cmd

import (
	"github.com/spf13/cobra"
)

var clusterDebugTerraformCmd = &cobra.Command{
	Use:   "terraform",
	Short: "terraform debug operations on clusters",
}

func init() {
	clusterDebugCmd.AddCommand(clusterDebugTerraformCmd)
}
