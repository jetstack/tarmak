// Copyright Jetstack Ltd. See LICENSE for details.
package cmd

import (
	"errors"

	"github.com/spf13/cobra"

	"github.com/jetstack/tarmak/pkg/tarmak"
)

var clusterImagesDestroyCmd = &cobra.Command{
	Use:   "destroy [image ids]",
	Short: "destroy remote tarmak images",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 && !globalFlags.Cluster.Images.Destroy.All {
			return errors.New("expecting at least a single image ID argument or --all")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		t := tarmak.New(globalFlags)
		t.CancellationContext().WaitOrCancel(t.NewCmdTarmak(cmd.Flags(), args).ImagesDestroy)
	},
}

func init() {
	clusterImagesDestroyFlags(clusterImagesDestroyCmd.PersistentFlags())
	clusterImagesCmd.AddCommand(clusterImagesDestroyCmd)
}
