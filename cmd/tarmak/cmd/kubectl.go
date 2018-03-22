// Copyright Jetstack Ltd. See LICENSE for details.
package cmd

import ()

func init() {
	RootCmd.AddCommand(clusterKubectlCmd)
}
