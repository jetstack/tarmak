// Copyright Jetstack Ltd. See LICENSE for details.
package cmd

import (
	"github.com/spf13/cobra"

	"github.com/jetstack/tarmak/pkg/tarmak"
)

var clusterDebugTerraformGenerateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate terraform code for current cluster",
	Run: func(cmd *cobra.Command, args []string) {
		t := tarmak.New(globalFlags)
		t.CancellationContext().WaitOrCancel(t.NewCmdTerraform(args).Generate)
	},
}

func init() {
	clusterDebugTerraformCmd.AddCommand(clusterDebugTerraformGenerateCmd)
}
