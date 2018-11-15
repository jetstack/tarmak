// Copyright Jetstack Ltd. See LICENSE for details.
package cmd

import (
	"github.com/spf13/cobra"

	"github.com/jetstack/tarmak/pkg/tarmak"
)

var clusterKubeconfigCmd = &cobra.Command{
	Use:   "kubeconfig",
	Short: "Verify and print path to Kubeconfig",
	Run: func(cmd *cobra.Command, args []string) {
		t := tarmak.New(globalFlags)
		t.Perform(t.NewCmdTarmak(args).Kubeconfig())
	},
}

func init() {
	clusterKubeconfigFlags(clusterKubeconfigCmd.PersistentFlags())
	clusterCmd.AddCommand(clusterKubeconfigCmd)
}
