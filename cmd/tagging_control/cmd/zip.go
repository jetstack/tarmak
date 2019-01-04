// Copyright Jetstack Ltd. See LICENSE for details.
package cmd

import (
	"errors"
	"fmt"

	"github.com/jetstack/tarmak/pkg/tarmak/utils/zip"
)

func Zip(args []string) error {

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
}
