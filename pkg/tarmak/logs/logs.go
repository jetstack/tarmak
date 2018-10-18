// Copyright Jetstack Ltd. See LICENSE for details.
package logs

import (
	"bytes"
	"encoding/json"
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

type SystemdEntry struct {
	Cursor                  string `json:"__CURSOR"`
	Realtime_timestamp      int64  `json:"__REALTIME_TIMESTAMP,string"`
	Monotonic_timestamp     string `json:"__MONOTONIC_TIMESTAMP"`
	Boot_id                 string `json:"_BOOT_ID"`
	Transport               string `json:"_TRANSPORT"`
	Priority                int32  `json:"PRIORITY,string"`
	SyslogFacility          string `json:"SYSLOG_FACILITY"`
	SyslogIdentifier        string `json:"SYSLOG_IDENTIFIER"`
	Pid                     string `json:"_PID"`
	Uid                     string `json:"_UID"`
	Gid                     string `json:"_GID"`
	Comm                    string `json:"_COMM"`
	Exe                     string `json:"_EXE"`
	Cmdline                 string `json:"_CMDLINE"`
	SystemdCGroup           string `json:"_SYSTEMD_CGROUP"`
	SystemdSession          string `json:"_SYSTEMD_SESSION"`
	SystemdOwnerUID         string `json:"_SYSTEMD_OWNER_UID"`
	SystemdUnit             string `json:"_SYSTEMD_UNIT"`
	SourceRealtimeTimestamp string `json:"_SOURCE_REALTIME_TIMESTAMP"`
	MachineID               string `json:"_MACHINE_ID"`
	Hostname                string `json:"_HOSTNAME"`

	// We require interface here since Message is not always a string. In the
	// case of vault-assets.service for example, some output is displayed as a
	// byte slice.
	Message interface{} `json:"MESSAGE"`
}

type instanceLogs struct {
	host        string
	serviceLogs map[string][]*SystemdEntry
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
		raw, err := l.fetchCmdOutput(name, "journalctl", []string{"-o", "json", "--no-pager"})
		if err != nil {
			return err
		}

		serviceLogs := make(map[string][]*SystemdEntry)
		for _, r := range bytes.Split(raw, []byte("\n")) {
			if r == nil || len(r) == 0 {
				continue
			}

			entry := new(SystemdEntry)
			if err := json.Unmarshal(r, entry); err != nil {
				return fmt.Errorf("failed to unmarshal entry [%s]: %s", r, err)
			}

			// ignore non-services
			if entry.SystemdUnit == "" {
				continue
			}

			serviceLogs[entry.SystemdUnit] = append(serviceLogs[entry.SystemdUnit], entry)
		}

		poolLogs = append(poolLogs, &instanceLogs{name, serviceLogs})

		select {
		case <-l.ctx.Done():
			return l.ctx.Err()
		default:
		}
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
		idir := filepath.Join(dir, i.host)
		if err := os.Mkdir(idir, os.FileMode(0755)); err != nil {
			result = multierror.Append(result, err)
			continue
		}

		var fileData []byte
		for unit, entries := range i.serviceLogs {
			for _, entry := range entries {
				fileData = append(
					fileData,
					[]byte(fmt.Sprintf("%s %s %s %s\n", entry.Realtime_timestamp, entry.Hostname, entry.SystemdUnit, entry.Message))...,
				)
			}

			ufile := filepath.Join(idir, unit)
			f, err := os.Create(ufile)
			if err != nil {
				result = multierror.Append(result, err)
				continue
			}
			defer f.Close()

			if _, err := f.Write(fileData); err != nil {
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
