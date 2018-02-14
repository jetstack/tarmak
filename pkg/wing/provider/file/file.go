// Copyright Jetstack Ltd. See LICENSE for details.
package file

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
)

type File struct {
	log *logrus.Entry
}

func New(log *logrus.Entry) *File {
	return &File{
		log: log,
	}
}

func (f *File) GetManifest(manifestURL string) (io.ReadCloser, error) {
	path := filepath.Join(manifestURL)
	fileStream, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("error opening file %s: %s", path, err)
	}
	return fileStream, nil
}
