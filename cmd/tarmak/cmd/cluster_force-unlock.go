// Copyright Jetstack Ltd. See LICENSE for details.
package cmd

import (
	"github.com/spf13/cobra"

	"github.com/jetstack/tarmak/pkg/tarmak"
)

var clusterForceUnlockCmd = &cobra.Command{
	Use:   "force-unlock",
	Short: "Remove remote lock using lock ID",
	Run: func(cmd *cobra.Command, args []string) {
		t := tarmak.New(globalFlags)
		defer t.Cleanup()

		forceUnlockCmd := t.NewCmdTerraform(args)

		t.CancellationContext().WaitOrCancel(forceUnlockCmd.ForceUnlock)
	},
}

func init() {
	clusterCmd.AddCommand(clusterForceUnlockCmd)
}
