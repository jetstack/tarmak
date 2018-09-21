// Copyright Jetstack Ltd. See LICENSE for details.
package cmd

import (
	"github.com/spf13/cobra"

	"github.com/jetstack/tarmak/pkg/tarmak"
)

var clusterImagesBuildCmd = &cobra.Command{
	Use:   "build",
	Short: "build images",
	Run: func(cmd *cobra.Command, args []string) {
		t := tarmak.New(globalFlags)
		t.CancellationContext().WaitOrCancel(t.Packer().Build)
	},
}

func init() {
	clusterImagesCmd.AddCommand(clusterImagesBuildCmd)
}
