// Copyright Jetstack Ltd. See LICENSE for details.
package cmd

import (
	"github.com/spf13/cobra"

	"github.com/jetstack/tarmak/pkg/tarmak"
)

// clusterLogsCmd handles `tarmak clusters logs`
var clusterLogsCmd = &cobra.Command{
	Use:   "logs",
	Short: "Gather logs from an instance pool",
	Run: func(cmd *cobra.Command, args []string) {
		t := tarmak.New(globalFlags)
		defer t.Cleanup()

		if len(args) == 0 {
			t.Log().Fatal("expecting at least a one instance pool name")
		}

		t.CancellationContext().WaitOrCancel(t.NewCmdTarmak(args).Logs)
	},
}

func init() {
	clusterLogsFlags(clusterLogsCmd.PersistentFlags())
	clusterCmd.AddCommand(clusterLogsCmd)
}
