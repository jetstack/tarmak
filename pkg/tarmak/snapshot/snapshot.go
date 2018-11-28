// Copyright Jetstack Ltd. See LICENSE for details.
package snapshot

import (
	"bufio"
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
	GZipCCmd   = "gzip -c %s"
	GZipDCmd   = "gzip -d > %s"
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

func TarFromStream(sshCmd func() error, stream io.Reader, path string) error {
	var result *multierror.Error
	var errLock sync.Mutex
	var wg sync.WaitGroup

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	wg.Add(1)
	go func() {
		defer wg.Done()
		err = sshCmd()
		if err != nil {

			errLock.Lock()
			result = multierror.Append(result, err)
			errLock.Unlock()

		}
		return
	}()

	gzr, err := gzip.NewReader(stream)
	if err != nil {

		errLock.Lock()
		result = multierror.Append(result, err)
		errLock.Unlock()

	}
	defer gzr.Close()

	if _, err := io.Copy(f, gzr); err != nil {

		errLock.Lock()
		result = multierror.Append(result, err)
		errLock.Unlock()

	}

	wg.Wait()

	if result != nil {
		return result
	}

	return nil
}

func TarToStream(sshCmd func() error, stream io.WriteCloser, src string) error {
	var result *multierror.Error
	var errLock sync.Mutex
	var wg sync.WaitGroup

	f, err := os.Open(src)
	if err != nil {
		return err
	}
	defer f.Close()

	wg.Add(1)
	go func() {
		defer wg.Done()
		err = sshCmd()
		if err != nil {

			errLock.Lock()
			result = multierror.Append(result, err)
			errLock.Unlock()

		}
		return
	}()

	gzw := gzip.NewWriter(stream)
	defer gzw.Close()
	if _, err := io.Copy(gzw, f); err != nil {

		errLock.Lock()
		result = multierror.Append(result, err)
		errLock.Unlock()

	}

	f.Close()
	gzw.Close()
	stream.Close()
	wg.Wait()

	if result != nil {
		return result
	}

	return nil
}

func SSHCmd(s interfaces.Snapshot, host, cmd string, stdin io.Reader, stdout, stderr io.Writer) error {
	fmt.Printf("$ %s\n", cmd)

	for _, w := range []struct {
		writer io.Writer
		out    string
	}{
		{stdout, "out"},
		{stderr, "err"},
	} {

		if w.writer == nil {
			var reader *io.PipeReader
			reader, w.writer = io.Pipe()
			scanner := bufio.NewScanner(reader)

			go func() {
				for scanner.Scan() {
					s.Log().WithField("std", w.out).Warn(scanner.Text())
				}
			}()
		}
	}

	ret, err := s.SSH().Execute(host, cmd, stdin, stdout, stderr)
	if ret != 0 {
		return fmt.Errorf("command [%s] returned non-zero (%d): %s", cmd, ret, err)
	}

	return err
}
