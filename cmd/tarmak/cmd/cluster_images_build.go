// Copyright Jetstack Ltd. See LICENSE for details.
package cmd

import (
	"github.com/spf13/cobra"

	"github.com/jetstack/tarmak/pkg/tarmak"
)

var clusterImagesBuildCmd = &cobra.Command{
	Use:   "build [base names]",
	Short: "build specific or all images missing",
	Run: func(cmd *cobra.Command, args []string) {
		t := tarmak.New(globalFlags)
		t.CancellationContext().WaitOrCancel(t.NewCmdTarmak(cmd.Flags(), args).ImagesBuild)
	},
}

func init() {
	clusterImagesBuildFlags(clusterImagesBuildCmd.PersistentFlags())
	clusterImagesCmd.AddCommand(clusterImagesBuildCmd)
}
