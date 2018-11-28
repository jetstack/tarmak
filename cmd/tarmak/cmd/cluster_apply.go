// Copyright Jetstack Ltd. See LICENSE for details.
package cmd

import (
	"errors"

	"github.com/spf13/cobra"

	"github.com/jetstack/tarmak/pkg/tarmak"
)

var clusterApplyCmd = &cobra.Command{
	Use:   "apply",
	Short: "Create or update the currently configured cluster",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		store := &globalFlags.Cluster.Apply

		if store.DryRun {
			return errors.New("dry run is not yet supported")
		}

		if store.InfrastructureOnly && store.ConfigurationOnly {
			return errors.New("the flags --infrastructure-only and --configuration-only are mutually exclusive")
		}

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		t := tarmak.New(globalFlags)
		applyCmd := t.NewCmdTarmak(cmd.Flags(), args)
		t.CancellationContext().WaitOrCancel(applyCmd.Apply)
	},
}

func init() {
	clusterApplyFlags(clusterApplyCmd.PersistentFlags())
	clusterCmd.AddCommand(clusterApplyCmd)
}
