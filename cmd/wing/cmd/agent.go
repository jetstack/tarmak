// Copyright Jetstack Ltd. See LICENSE for details.
package cmd

import (
	"github.com/spf13/cobra"

	"github.com/jetstack/tarmak/pkg/apis/wing/common"
	"github.com/jetstack/tarmak/pkg/wing"
)

var agentFlags = &common.Flags{}

var agentCmd = &cobra.Command{
	Use: "agent",
	Run: func(cmd *cobra.Command, args []string) {
		w := wing.New(agentFlags)
		w.Must(w.Run(args))
	},
}

func init() {
	agentCmd.Flags().StringVar(&agentFlags.ClusterName, "cluster-name", "myenv-mycluster", "this specifies the cluster name [environment]-[cluster]")
	agentCmd.Flags().StringVar(&agentFlags.ServerURL, "server-url", "https://localhost:9443", "this specifies the URL to the wing server")
	agentCmd.Flags().StringVar(&agentFlags.ManifestURL, "manifest-url", "", "this specifies the URL where the puppet.tar.gz can be found")
	agentCmd.Flags().StringVar(&agentFlags.MachineName, "instance-name", wing.DefaultMachineName, "this specifies the instance's name")
	agentCmd.Flags().StringVar(&agentFlags.Role, "role", "", "this specifies the machines role, used as a label selector for machinesets")

	RootCmd.AddCommand(agentCmd)
}
