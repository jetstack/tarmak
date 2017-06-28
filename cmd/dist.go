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

	"github.com/mholt/archiver"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// distCmd represents the dist command
var distCmd = &cobra.Command{
	Use:   "dist",
	Short: "Build a puppet.tar.gz",
	Run: func(cmd *cobra.Command, args []string) {

		oldPwd, err := os.Getwd()
		if err != nil {
			log.Fatal("error getting pwd: ", err)
		}

		puppetDir := "./puppet"
		err = os.Chdir("puppet")
		if err != nil {
			log.Fatal("Error changing directory to '%s': %s", puppetDir, err)
		}

		err = archiver.TarGz.Make("../puppet.tar.gz", []string{"manifests", "modules", "hieradata", "hiera.yaml"})
		if err != nil {
			log.Fatal("Error creating puppet.tar.gz: ", err)
		}
		log.Info("created puppet.tar.gz")

		err = os.Chdir(oldPwd)
		if err != nil {
			log.Fatal("Error changing directory to '%s': %s", puppetDir, err)
		}
	},
}

func init() {
	RootCmd.AddCommand(distCmd)
}
