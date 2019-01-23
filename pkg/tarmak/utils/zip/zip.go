// Copyright Jetstack Ltd. See LICENSE for details.
package zip

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/jetstack/tarmak/pkg/tarmak/utils"
)

var (
	overrideErr = "trying to override source file but not deleting source: %s %s"
)

func Zip(src []string, dst string, deleteSrc bool) error {
	var writer io.Writer
	var tempFile *os.File

	if !strings.HasSuffix(dst, ".zip") {
		dst = fmt.Sprintf("%s.zip", dst)
	}

	if utils.SliceContains(src, dst) {
		if !deleteSrc {
			return fmt.Errorf(overrideErr, src, dst)
		}

		// source contains destination file so write zip to temp file
		var err error
		tempFile, err = os.Create(filepath.Join(os.TempDir(), filepath.Base(dst)))
		if err != nil {
			return err
		}

		defer os.RemoveAll(tempFile.Name())
		writer = tempFile

	} else {

		// source files doesn't contain destination so write straight to
		// destination file
		f, err := os.Create(dst)
		if err != nil {
			return err
		}

		defer f.Close()
		writer = f
	}

	zipW := zip.NewWriter(writer)
	defer zipW.Close()

	for _, f := range src {
		srcF, err := os.Open(f)
		if err != nil {
			return err
		}
		defer srcF.Close()

		info, err := srcF.Stat()
		if err != nil {
			return err
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		header.Name = f
		header.Method = zip.Deflate

		fWriter, err := zipW.CreateHeader(header)
		if err != nil {
			return err
		}

		if _, err = io.Copy(fWriter, srcF); err != nil {
			return err
		}
	}

	if deleteSrc {
		for _, f := range src {
			if err := os.RemoveAll(f); err != nil {
				return fmt.Errorf("failed to delete all src files: %s", err)
			}
		}
	}

	zipW.Close()

	// zip is in temp so copy to destination file
	if tempFile != nil {
		tempFile.Close()

		tempFile, err := os.Open(tempFile.Name())
		if err != nil {
			return err
		}

		f, err := os.Create(dst)
		if err != nil {
			return err
		}
		defer f.Close()

		if _, err = io.Copy(f, tempFile); err != nil {
			return err
		}
	}

	return nil
}
