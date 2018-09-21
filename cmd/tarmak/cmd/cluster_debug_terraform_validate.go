// Copyright Jetstack Ltd. See LICENSE for details.
package cmd

import (
	"github.com/spf13/cobra"

	"github.com/jetstack/tarmak/pkg/tarmak"
)

var clusterDebugTerraformValidateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate terraform code for current cluster",
	Run: func(cmd *cobra.Command, args []string) {
		t := tarmak.New(globalFlags)
		t.CancellationContext().WaitOrCancel(t.NewCmdTerraform(args).Validate)
	},
}

func init() {
	clusterDebugTerraformCmd.AddCommand(clusterDebugTerraformValidateCmd)
}
