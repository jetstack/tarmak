// Copyright Jetstack Ltd. See LICENSE for details.
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/jetstack/tarmak/pkg/tarmak"
	"github.com/jetstack/tarmak/pkg/tarmak/snapshot/consul"
)

var clusterSnapshotConsulSaveCmd = &cobra.Command{
	Use:   "save [target path]",
	Short: "save consul cluster snapshot to target path",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("expecting single target path, got=%d", len(args))
		}

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		t := tarmak.New(globalFlags)
		s := consul.New(t, args[0])
		t.CancellationContext().WaitOrCancel(t.NewCmdSnapshot(cmd.Flags(), args, s).Save)
	},
}

func init() {
	clusterSnapshotConsulCmd.AddCommand(clusterSnapshotConsulSaveCmd)
}
