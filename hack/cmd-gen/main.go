// Copyright Jetstack Ltd. See LICENSE for details.
package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mitchellh/go-homedir"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"

	tarmakCmd "github.com/jetstack/tarmak/cmd/tarmak/cmd"
	wingCmd "github.com/jetstack/tarmak/cmd/wing/cmd"
)

func main() {
	args := os.Args
	if len(args) != 2 {
		log.Fatal("expecting single output directory argument")
	}

	root, err := homedir.Expand(args[1])
	must(err)

	must(ensureDirectory(root))

	for _, c := range []*cobra.Command{
		tarmakCmd.RootCmd,
		wingCmd.RootCmd,
	} {
		dir := filepath.Join(root, c.Use)
		must(ensureDirectory(dir))
		must(doc.GenReSTTree(c, dir))
	}
}

func ensureDirectory(dir string) error {
	s, err := os.Stat(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return os.Mkdir(dir, os.FileMode(0755))
		}
		return err
	}

	if !s.IsDir() {
		return fmt.Errorf("path it not directory: %s", dir)
	}

	return nil
}

func must(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
