// Copyright Jetstack Ltd. See LICENSE for details.
package cmd

import (
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/jetstack/tarmak/pkg/tarmak"
	"github.com/jetstack/tarmak/pkg/tarmak/utils"
)

var clusterInstancesListCmd = &cobra.Command{
	Use:   "list",
	Short: "Print a list of instances in the cluster",
	Run: func(cmd *cobra.Command, args []string) {
		t := tarmak.New(globalFlags)
		defer t.Cleanup()
		hosts, err := t.Cluster().ListHosts()
		if err != nil {
			logrus.Fatal(err)
		}

		varMaps := make([]map[string]string, 0)
		for _, host := range hosts {
			varMaps = append(varMaps, host.Parameters())
		}
		utils.ListParameters(os.Stdout, []string{"id", "hostname", "roles"}, varMaps)
	},
}

func init() {
	clusterInstancesCmd.AddCommand(clusterInstancesListCmd)
}
