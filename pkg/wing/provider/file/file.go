// Copyright Jetstack Ltd. See LICENSE for details.
package file

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type File struct{}

func (f *File) GetManifest(manifestURL string) (io.ReadCloser, error) {
	path := filepath.Join(manifestURL)
	filestream, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("error opening file %s: %s", path, err)
	}
	return filestream, nil
}

func (f *File) Name() string {
	return "file"
}
