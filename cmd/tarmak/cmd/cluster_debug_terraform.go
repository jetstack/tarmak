// Copyright Jetstack Ltd. See LICENSE for details.
package cmd

import (
	"github.com/spf13/cobra"
)

var clusterDebugTerraformCmd = &cobra.Command{
	Use:   "terraform",
	Short: "Operations for debugging Terraform configuration",
}

func init() {
	clusterDebugCmd.AddCommand(clusterDebugTerraformCmd)
}
