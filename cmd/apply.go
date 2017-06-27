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
	"bufio"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/mholt/archiver"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var applyCmd = &cobra.Command{
	Use:   "apply",
	Short: "Apply a puppet.tar.gz locally",
	Run: func(cmd *cobra.Command, args []string) {
		log.SetLevel(log.DebugLevel)
		dir, err := ioutil.TempDir("", "tarmak-apply")
		if err != nil {
			log.Fatal(err)
		}
		defer os.RemoveAll(dir) // clean up

		err = archiver.TarGz.Open("puppet.tar.gz", dir)
		if err != nil {
			log.Fatal(err)
		}

		hieraData := `---
version: 5
defaults:
  datadir: data
  data_hash: yaml_data

hierarchy:
- name: Per node data
  path: "nodes/%{::trusted.certname}.yaml"
- name: Per role data
  path: "role/%{::tarmak_role}.yaml"
- name: Per environment data
  path: "environment/%{tarmak_environment}.yaml"
- name: Default fallback
  path: common.yaml
`
		err = ioutil.WriteFile(
			filepath.Join(dir, "hiera.yaml"),
			[]byte(hieraData),
			0644,
		)
		if err != nil {
			log.Fatal(err)
		}

		puppetCmd := exec.Command(
			"puppet",
			"apply",
			"--hiera_config",
			filepath.Join(dir, "hiera.yaml"),
			"--modulepath",
			filepath.Join(dir, "modules"),
			filepath.Join(dir, "manifests/site.pp"),
		)

		stdoutPipe, err := puppetCmd.StdoutPipe()
		if err != nil {
			log.Fatal(err)
		}

		stderrPipe, err := puppetCmd.StderrPipe()
		if err != nil {
			log.Fatal(err)
		}

		stdoutScanner := bufio.NewScanner(stdoutPipe)
		go func() {
			for stdoutScanner.Scan() {
				log.WithField("cmd", "puppet").Debug(stdoutScanner.Text())
			}
		}()

		stderrScanner := bufio.NewScanner(stderrPipe)
		go func() {
			for stderrScanner.Scan() {
				log.WithField("cmd", "puppet").Debug(stderrScanner.Text())
			}
		}()

		err = puppetCmd.Start()
		if err != nil {
			log.Fatal(err)
		}

		log.Printf("Waiting for command to finish...")
		err = puppetCmd.Wait()
		log.Printf("Command finished with error: %v", err)

	},
}

func init() {
	RootCmd.AddCommand(applyCmd)
}
