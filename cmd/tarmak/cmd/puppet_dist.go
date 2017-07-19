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
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/jetstack/tarmak/pkg/tarmak"
)

// distCmd represents the dist command
var puppetDistCmd = &cobra.Command{
	Use:   "puppet-dist",
	Short: "Build a puppet.tar.gz",
	Run: func(cmd *cobra.Command, args []string) {
		t := tarmak.New(cmd)

		path := "puppet.tar.gz"

		file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0600)
		if err != nil {
			log.Fatalf("error creating %s: %s", path, err)
		}

		if err = t.Puppet().TarGz(file); err != nil {
			log.Fatalf("error writing to %s: %s", path, err)
		}

		if err := file.Close(); err != nil {
			log.Fatalf("error closing %s: %s", path, err)
		}
	},
}

func init() {
	RootCmd.AddCommand(puppetDistCmd)
}
