// Copyright Jetstack Ltd. See LICENSE for details.
package cmd

import (
	"errors"
	"fmt"
	"github.com/spf13/cobra"

	"github.com/jetstack/tarmak/pkg/tarmak/utils/zip"
)

var zipCmd = &cobra.Command{
	Use:   "zip [dst] [src...]",
	Short: "zip utility",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 2 {
			return errors.New("zip expecting files: [dst] [src...]")
		}

		dst := args[0]
		src := args[1:]

		fmt.Printf("deflating %s into %s...\n", src, dst)

		if err := zip.Zip(src, dst, true); err != nil {
			return fmt.Errorf("zip failed: %s", err)
		}

		fmt.Print("done.\n")

		return nil
	},
}

func init() {
	RootCmd.AddCommand(zipCmd)
}
