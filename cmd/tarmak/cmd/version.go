// Copyright Jetstack Ltd. See LICENSE for details.
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	pkgversion "github.com/jetstack/tarmak/pkg/version"
)

var AppName string = "tarmak"

func init() {
	RootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: fmt.Sprint("Print the version number of ", AppName),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("%s version: %#v\n", AppName, pkgversion.Get())
	},
}
