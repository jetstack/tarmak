// Copyright Jetstack Ltd. See LICENSE for details.
package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"
	"time"

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

		format := "%s\t%s\t%s\t%s\t%s\n"
		fmt.Fprintf(
			w,
			format,
			"Image ID",
			"Base Image",
			"Location",
			"Tags",
			"Created",
		)

		for _, image := range images {
			fmt.Fprintf(
				w,
				format,
				image.Name,
				image.BaseImage,
				image.Location,
				image.Annotations,
				image.CreationTimestamp.Format(time.RFC3339),
			)
		}
		w.Flush()
	},
}

func init() {
	clusterImagesCmd.AddCommand(clusterImagesListCmd)
}
