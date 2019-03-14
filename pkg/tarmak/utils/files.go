// Copyright Jetstack Ltd. See LICENSE for details.
package utils

import (
	"path/filepath"

	"github.com/mitchellh/go-homedir"
)

func Expand(path string) (string, error) {
	p, err := homedir.Expand(path)
	if err != nil {
		return "", err
	}

	return filepath.Abs(p)
}
