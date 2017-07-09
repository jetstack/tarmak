package utils

import (
	"os"
	"path/filepath"
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
