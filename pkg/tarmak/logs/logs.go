// Copyright Jetstack Ltd. See LICENSE for details.
package logs

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/docker/docker/pkg/archive"
	"github.com/hashicorp/go-multierror"
	"github.com/mitchellh/go-homedir"
	"github.com/sirupsen/logrus"

	"github.com/jetstack/tarmak/pkg/tarmak/interfaces"
	"github.com/jetstack/tarmak/pkg/tarmak/utils"
)

type Logs struct {
	tarmak       interfaces.Tarmak
	ssh          interfaces.SSH
	instancePool interfaces.InstancePool
	ctx          interfaces.CancellationContext
	log          *logrus.Entry

	path string
}

type instanceLogs struct {
	name         string
	servicesLogs []*serviceLogs
}
type serviceLogs struct {
	name string
	logs []byte
}

func New(tarmak interfaces.Tarmak) *Logs {
	return &Logs{
		tarmak: tarmak,
		log:    tarmak.Log(),
		ctx:    tarmak.CancellationContext(),
	}
}

func (l *Logs) Gather(pool, path string) error {
	if i := l.tarmak.Cluster().InstancePool(pool); i == nil {
		return fmt.Errorf("unable to find instance pool '%s'", pool)
	}

	if err := l.setPath(pool, path); err != nil {
		return fmt.Errorf("failed to set tar gz log bundle path: %v", err)
	}

	l.ssh = l.tarmak.SSH()

	err := l.ssh.WriteConfig(l.tarmak.Cluster())
	if err != nil {
		return err
	}

	hosts, err := l.tarmak.Cluster().ListHosts()
	if err != nil {
		return err
	}

	var poolLogs []*instanceLogs
	for _, host := range hosts {
		var name string
		for _, r := range host.Roles() {
			if strings.HasPrefix(r, pool) {
				name = r
				break
			}
		}

		if name == "" {
			continue
		}

		l.log.Infof("fetching service logs from instance '%s'", name)
		services, err := l.listServices(name)
		if err != nil {
			l.log.Errorf("failed to gather unit service list from instance '%s', skipping...", name)
			continue
		}

		var servicesLogs []*serviceLogs
		for _, s := range services {

			select {
			case <-l.ctx.Done():
				return l.ctx.Err()
			default:
			}

			if s != "" {
				l.log.Debugf("fetching logs [%s]: %s", name, s)
				stdout, err := l.fetchServiceJournal(name, s)
				if err != nil {
					l.log.Errorf("failed to gather logs [%s]:%s, skipping... %v", name, s, err)
				}

				servicesLogs = append(servicesLogs, &serviceLogs{s, stdout})
			}
		}

		poolLogs = append(poolLogs, &instanceLogs{name, servicesLogs})
	}

	select {
	case <-l.ctx.Done():
		return l.ctx.Err()
	default:
	}

	return l.bundleLogs(poolLogs)
}

func (l *Logs) bundleLogs(poolLogs []*instanceLogs) error {
	dir, err := ioutil.TempDir("", filepath.Base(l.path))
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)

	var result *multierror.Error
	for _, i := range poolLogs {
		idir := filepath.Join(dir, i.name)
		if err := os.Mkdir(idir, os.FileMode(0755)); err != nil {
			result = multierror.Append(result, err)
			continue
		}

		for _, s := range i.servicesLogs {
			sfile := filepath.Join(idir, s.name)
			f, err := os.Create(sfile)
			if err != nil {
				result = multierror.Append(result, err)
				continue
			}
			defer f.Close()

			if _, err := f.Write(s.logs); err != nil {
				result = multierror.Append(result, err)
			}
		}
	}

	if result != nil {
		return result.ErrorOrNil()
	}

	f, err := os.Create(l.path)
	if err != nil {
		return err
	}
	defer f.Close()

	reader, err := archive.Tar(
		dir,
		archive.Gzip,
	)
	if err != nil {
		return fmt.Errorf("error creating tar from path '%s': %s", l.path, err)
	}

	if _, err := io.Copy(f, reader); err != nil {
		return fmt.Errorf("error writing tar: %s", err)
	}

	l.log.Infof("logs bundle written to '%s'", f.Name())

	return nil
}

func (l *Logs) listServices(host string) ([]string, error) {
	args := "list-unit-files --type=service --no-pager | tail -n +2 | head -n -1 | awk '{print $1}'"
	stdout, err := l.fetchCmdOutput(host, "systemctl", strings.Split(args, " "))

	return strings.Split(string(stdout), "\n"), err
}

func (l *Logs) fetchServiceJournal(host, service string) ([]byte, error) {
	return l.fetchCmdOutput(host, "journalctl", []string{"-u", service, "--no-pager"})
}

func (l *Logs) fetchCmdOutput(host, command string, args []string) ([]byte, error) {
	var stdout bytes.Buffer
	ret, err := l.ssh.ExecuteWithWriter(host, command, args, &stdout)
	if ret != 0 {
		return nil, fmt.Errorf("command returned non-zero (%d): %v", ret, err)
	}
	return stdout.Bytes(), err
}

func (l *Logs) setPath(pool, path string) error {
	if path == utils.DefaultLogsPathPlaceholder {
		l.path = filepath.Join(l.tarmak.Cluster().ConfigPath(), fmt.Sprintf("%s-logs.tar.gz", pool))
		return nil
	}

	p, err := homedir.Expand(path)
	if err != nil {
		return err
	}

	p, err = filepath.Abs(p)
	if err != nil {
		return err
	}
	l.path = p

	return nil
}
