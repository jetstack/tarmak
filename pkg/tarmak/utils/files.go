// Copyright Jetstack Ltd. See LICENSE for details.
package utils

import (
	"os"
	"path/filepath"

	"github.com/mitchellh/go-homedir"
)

func EnsureDirectory(path string, mode os.FileMode) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := EnsureDirectory(filepath.Dir(path), mode); err != nil {
			return err
		}
		os.Mkdir(path, mode)
	} else {
		return err
	}
	return nil
}

func Expand(path string) (string, error) {
	p, err := homedir.Expand(path)
	if err != nil {
		return "", err
	}

	return filepath.Abs(p)
}
