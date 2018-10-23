// Copyright Jetstack Ltd. See LICENSE for details.
package cmd

import (
	"github.com/spf13/cobra"

	"github.com/jetstack/tarmak/pkg/tarmak"
)

var clusterForceUnlockCmd = &cobra.Command{
	Use:   "force-unlock [lock ID]",
	Short: "Remove remote lock using lock ID",
	Run: func(cmd *cobra.Command, args []string) {
		t := tarmak.New(globalFlags)

		t.CancellationContext().WaitOrCancel(t.NewCmdTerraform(args).ForceUnlock)
	},
}

func init() {
	clusterCmd.AddCommand(clusterForceUnlockCmd)
}
