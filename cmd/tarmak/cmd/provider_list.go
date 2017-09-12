package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/jetstack/tarmak/pkg/tarmak"
	"github.com/jetstack/tarmak/pkg/tarmak/provider"
)

var providerListCmd = &cobra.Command{
	Use:   "list",
	Short: "list providers",
	Run: func(cmd *cobra.Command, args []string) {
		t := tarmak.New(cmd)

		w := new(tabwriter.Writer)
		w.Init(os.Stdout, 0, 8, 0, '\t', 0)

		fmt.Fprintf(
			w,
			"%s\t%s\t%s\n",
			"Name",
			"Provider",
			"Parameters",
		)

		for _, providerConf := range t.Config().Providers() {
			providerObj, err := provider.NewProviderFromConfig(t, providerConf)
			if err != nil {
				t.Log().Warn("error listing provider '%s': %s", providerConf.Name, err)
				continue
			}

			fmt.Fprintf(
				w,
				"%s\t%s\t%+v\n",
				providerConf.Name,
				providerObj.Name(),
				providerObj.Parameters(),
			)
		}
		w.Flush()
	},
}

func init() {
	providerCmd.AddCommand(providerListCmd)
}
