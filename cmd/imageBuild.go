package cmd

import (
	"github.com/spf13/cobra"

	"github.com/jetstack/tarmak/pkg/tarmak"
)

// tfapplyCmd represents the tfapply command
var imageBuildCmd = &cobra.Command{
	Use:   "image-build",
	Short: "This builds an image for an environment using packer",
	Run: func(cmd *cobra.Command, args []string) {
		t := tarmak.New()
		t.PackerBuild()
	},
}

func init() {
	RootCmd.AddCommand(imageBuildCmd)
}
