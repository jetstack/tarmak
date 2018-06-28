// Copyright Jetstack Ltd. See LICENSE for details.
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/jetstack/tarmak/pkg/tarmak"
)

var clusterKubeconfigCmd = &cobra.Command{
	Use:   "kubeconfig",
	Short: "Verify and get path to Kubeconfig",
	Run: func(cmd *cobra.Command, args []string) {
		t := tarmak.New(globalFlags)
		defer t.Cleanup()
		kubeconfig, error := t.CmdKubeconfig(cmd.Flags())

		t.Must(error)
		fmt.Printf("%s", kubeconfig)
	},
}

func init() {
	clusterKubeconfigFlags(clusterKubeconfigCmd.PersistentFlags())
	clusterCmd.AddCommand(clusterKubeconfigCmd)
}