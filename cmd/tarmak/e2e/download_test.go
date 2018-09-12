// Copyright Jetstack Ltd. See LICENSE for details.
package e2e_test

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
)

func EnsureTarmakDownload(dir, version string) error {
	bin := fmt.Sprintf("tarmak_%s_%s_%s",
		version, runtime.GOOS, runtime.GOARCH)
	binPath := filepath.Join(dir, bin)

	stat, err := os.Stat(binPath)
	if err == nil {
		if stat.IsDir() {
			return fmt.Errorf("target tarmak binary path is directory %s", binPath)
		}

		return nil
	} else if !os.IsNotExist(err) {
		return err
	}

	url := fmt.Sprintf("https://www.github.com/jetstack/tarmak/releases/download/%s/%s",
		version, bin)

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	f, err := os.OpenFile(binPath, os.O_CREATE|os.O_WRONLY, 0744)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, resp.Body)
	return err
}
