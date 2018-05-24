// Copyright Jetstack Ltd. See LICENSE for details.
package cmd

var (
	version = "dev"
)

func init() {
	RootCmd.AddCommand(clusterInitCmd)
}
