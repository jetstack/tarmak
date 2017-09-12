package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/jetstack/tarmak/pkg/tarmak"
)

var clusterImagesListCmd = &cobra.Command{
	Use:   "list",
	Short: "list images",
	Run: func(cmd *cobra.Command, args []string) {
		t := tarmak.New(cmd)

		w := new(tabwriter.Writer)
		w.Init(os.Stdout, 0, 8, 0, '\t', 0)

		images, err := t.Packer().List()
		t.Must(err)

		fmt.Fprintf(
			w,
			"%s\t%s\t%s\t%s\n",
			"Image ID",
			"Base Image",
			"Location",
			"Tags",
		)

		for _, image := range images {
			fmt.Fprintf(
				w,
				"%s\t%s\t%s\t%s\n",
				image.Name,
				image.BaseImage,
				image.Location,
				image.Annotations,
			)
		}
		w.Flush()
	},
}

func init() {
	clusterImagesCmd.AddCommand(clusterImagesListCmd)
}
