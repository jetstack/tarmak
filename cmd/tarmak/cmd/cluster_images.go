package cmd

import (
	"github.com/spf13/cobra"
)

var clusterImagesCmd = &cobra.Command{
	Use:     "images",
	Short:   "Operations on images",
	Aliases: []string{"image"},
}

func init() {
	clusterCmd.AddCommand(clusterImagesCmd)
}
