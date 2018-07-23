// Copyright Jetstack Ltd. See LICENSE for details.
package cmd

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/jetstack/tarmak/pkg/tarmak"
	"github.com/jetstack/tarmak/pkg/tarmak/utils"
)

var clusterImagesBuildCmd = &cobra.Command{
	Use:   "build",
	Short: "build images",
	Run: func(cmd *cobra.Command, args []string) {
		t := tarmak.New(globalFlags)
		defer t.Cleanup()
		utils.WaitOrCancel(
			func(ctx context.Context) error {
				return t.Packer().Build(ctx)
			},
			t.Context(),
		)
	},
}

func init() {
	clusterImagesCmd.AddCommand(clusterImagesBuildCmd)
}
