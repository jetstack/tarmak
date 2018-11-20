// Copyright Jetstack Ltd. See LICENSE for details.
package cmd

func init() {
	clusterInstancesCmd.AddCommand(clusterSshCmd)
}
