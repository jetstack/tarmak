package docker

import (
	"archive/tar"
	"bytes"
	"fmt"
	"io"
)

func TarStreamFromDockerfile(fileBody string) (io.Reader, error) {
	return TarStreamFromFile("Dockerfile", fileBody)
}

func TarStreamFromFile(fileName, fileBody string) (io.Reader, error) {
	buf := new(bytes.Buffer)

	t := tar.NewWriter(buf)

	hdr := &tar.Header{
		Name: fileName,
		Mode: 0644,
		Size: int64(len(fileBody)),
	}
	if err := t.WriteHeader(hdr); err != nil {
		return nil, fmt.Errorf("error writing file header: %s", err)
	}
	if _, err := t.Write([]byte(fileBody)); err != nil {
		return nil, fmt.Errorf("error writing file body: %s", err)
	}
	if err := t.Close(); err != nil {
		return nil, fmt.Errorf("error closing tar: %s", err)
	}

	return bytes.NewReader(buf.Bytes()), nil
}
