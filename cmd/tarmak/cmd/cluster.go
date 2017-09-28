package cmd

import (
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"

	"github.com/jetstack/tarmak/pkg/tarmak"
)

var clusterCmd = &cobra.Command{
	Use:     "clusters",
	Short:   "operations on clusters",
	Aliases: []string{"cluster"},
}

func terraformPFlags(fs *flag.FlagSet) {
	fs.StringSlice(tarmak.FlagTerraformStacks, []string{}, "terraform stacks to execute")
}

func init() {
	RootCmd.AddCommand(clusterCmd)
}
