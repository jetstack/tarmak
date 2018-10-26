// Copyright Jetstack Ltd. See LICENSE for details.
package cmd

import (
	"errors"

	"github.com/spf13/cobra"

	"github.com/jetstack/tarmak/pkg/tarmak"
)

var clusterImagesDestroyCmd = &cobra.Command{
	Use:   "destroy [image ids]",
	Short: "destroy images",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return errors.New("expecting at least a single image ID argument")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		t := tarmak.New(globalFlags)
		t.CancellationContext().WaitOrCancel(t.NewCmdTarmak(args).ImagesDestroy)
	},
}

func init() {
	clusterImagesCmd.AddCommand(clusterImagesDestroyCmd)
}
