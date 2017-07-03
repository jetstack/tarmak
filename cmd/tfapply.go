// Copyright © 2017 NAME HERE <EMAIL ADDRESS>
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	log "github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/jetstack/tarmak/pkg/tarmak/config"
	"github.com/jetstack/tarmak/pkg/terraform"
)

// tfapplyCmd represents the tfapply command
var tfapplyCmd = &cobra.Command{
	Use:   "tfapply",
	Short: "This applies the set of stacks in the current context",
	Run: func(cmd *cobra.Command, args []string) {
		log.Info("tfapply called")

		myConfig := config.DefaultConfigSingle()

		err := myConfig.Validate()
		if err != nil {
			log.Fatal(err)
		}

		context, err := myConfig.GetContext()
		if err != nil {
			log.Fatal(err)
		}
		log.Debugf("current context: %#+v", context)

		tf := terraform.New(nil, context)

		for posStack, _ := range context.Stacks {
			err = tf.Apply(&context.Stacks[posStack])
			if err != nil {
				log.Fatal(err)
			}
		}

	},
}

func init() {
	RootCmd.AddCommand(tfapplyCmd)
}
