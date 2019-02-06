// Copyright Jetstack Ltd. See LICENSE for details.
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/jetstack/tarmak/pkg/tarmak"
	"github.com/jetstack/tarmak/pkg/tarmak/logs"
)

// clusterLogsCmd handles `tarmak clusters logs`
var clusterLogsCmd = &cobra.Command{
	Use: "logs [target groups]",
	Long: fmt.Sprintf(
		"Gather logs from a list of instances or target groups %s",
		logs.TargetGroups,
	),
	Aliases: []string{"log"},
	Short:   "Gather logs from a list of instances or target groups",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf(
				"expecting at least one instance or target group %s",
				logs.TargetGroups,
			)
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		t := tarmak.New(globalFlags)
		defer t.Cleanup()
		t.CancellationContext().WaitOrCancel(t.NewCmdTarmak(cmd.Flags(), args).Logs)
	},
}

func init() {
	clusterLogsFlags(clusterLogsCmd.PersistentFlags())
	clusterCmd.AddCommand(clusterLogsCmd)
}
