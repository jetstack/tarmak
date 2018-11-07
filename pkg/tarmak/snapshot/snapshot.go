// Copyright Jetstack Ltd. See LICENSE for details.
package snapshot

import (
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/hashicorp/go-multierror"

	"github.com/jetstack/tarmak/pkg/tarmak/interfaces"
	"github.com/jetstack/tarmak/pkg/tarmak/utils"
)

const (
	TimeLayout = "2006-01-02_15-04-05"
	TarCCmd    = "tar -czPf - %s"
	//TarXCmd    = "cat | tar -xz | cat > /tmp/foo" //| cat > %s"
	TarXCmd = "cat > %s"
)

func Prepare(tarmak interfaces.Tarmak, role string) (aliases []string, err error) {
	if err := tarmak.SSH().WriteConfig(tarmak.Cluster()); err != nil {
		return nil, err
	}

	hosts, err := tarmak.Cluster().ListHosts()
	if err != nil {
		return nil, err
	}

	var result *multierror.Error
	for _, host := range hosts {
		if utils.SliceContainsPrefix(host.Roles(), role) {
			if len(host.Aliases()) == 0 {
				err := fmt.Errorf(
					"host with correct role '%v' found without alias: %v",
					host.Roles(),
					host.ID(),
				)
				result = multierror.Append(result, err)
				continue
			}

			aliases = append(aliases, host.Aliases()[0])
		}
	}

	if result != nil {
		return nil, result
	}

	if len(aliases) == 0 {
		return nil, fmt.Errorf("no host aliases were found with role %s", role)
	}

	return aliases, result.ErrorOrNil()
}

func ReadTarFromStream(dest string, stream io.Reader, result *multierror.Error, errLock sync.Mutex) {
	gzr, err := gzip.NewReader(stream)
	if err != nil {

		errLock.Lock()
		result = multierror.Append(result, err)
		errLock.Unlock()

		return
	}

	f, err := os.Create(dest)
	if err != nil {

		errLock.Lock()
		result = multierror.Append(result, err)
		errLock.Unlock()

		return
	}

	if _, err := io.Copy(f, gzr); err != nil {

		errLock.Lock()
		result = multierror.Append(result, err)
		errLock.Unlock()

		return
	}
}

func WriteTarToStream(src string, stream io.WriteCloser, result *multierror.Error, errLock sync.Mutex) {
	//defer stream.Close()

	f, err := os.Open(src)
	if err != nil {

		errLock.Lock()
		result = multierror.Append(result, err)
		errLock.Unlock()

		return
	}

	gzw := gzip.NewWriter(stream)
	if _, err := io.Copy(gzw, f); err != nil {

		errLock.Lock()
		result = multierror.Append(result, err)
		errLock.Unlock()

		return
	}
}
