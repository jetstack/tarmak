// Copyright Jetstack Ltd. See LICENSE for details.
package zip

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"time"
)

var (
	overrideErr = "trying to override source file but not deleting source: %s %s"
)

func ZipBytes(filenames []string, bytes [][]byte, modes []os.FileMode, writer io.Writer) error {
	zipW := zip.NewWriter(writer)
	defer zipW.Close()

	kubernetesEpoch := time.Unix(1437436800, 0)

	if lenFilenames, lenBytes, lenModes := len(filenames), len(bytes), len(modes); lenFilenames != lenBytes || lenBytes != lenModes {
		return fmt.Errorf("count of filenames, modes and bytes slice in the input needs to match")
	}

	for pos, _ := range filenames {
		header := &zip.FileHeader{
			Name:               filenames[pos],
			Method:             zip.Deflate,
			UncompressedSize64: uint64(len(bytes[pos])),
			Modified:           time.Now(),
		}
		header.SetMode(modes[pos])
		header.SetModTime(kubernetesEpoch)

		fWriter, err := zipW.CreateHeader(header)
		if err != nil {
			return err
		}
		if _, err := fWriter.Write(bytes[pos]); err != nil {
			return err
		}
	}
	return nil
}
