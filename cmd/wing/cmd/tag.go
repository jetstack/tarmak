// Copyright Jetstack Ltd. See LICENSE for details.
package cmd

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/jetstack/tarmak/pkg/wing/tags"
)

var tagCmd = &cobra.Command{
	Use:   "tag",
	Short: "Ensure public key tagging of instance",
	Run: func(cmd *cobra.Command, args []string) {
		log := logrus.New()
		log.SetLevel(logrus.DebugLevel)

		env, err := cmd.Flags().GetString("environment")
		if err != nil {
			log.Fatal(err)
		}

		t, err := tags.New(logrus.NewEntry(log), env)
		if err != nil {
			log.Fatal(err)
		}

		if err := t.EnsureMachineTags(); err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	tagCmd.Flags().String("environment", "", "this specifies the environment name")
	RootCmd.AddCommand(tagCmd)
}
