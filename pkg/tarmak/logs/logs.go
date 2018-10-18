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
	"sync"
	"time"

	"github.com/docker/docker/pkg/archive"
	"github.com/hashicorp/go-multierror"
	"github.com/mitchellh/go-homedir"
	"github.com/sirupsen/logrus"

	tarmakv1alpha1 "github.com/jetstack/tarmak/pkg/apis/tarmak/v1alpha1"
	"github.com/jetstack/tarmak/pkg/tarmak/interfaces"
	"github.com/jetstack/tarmak/pkg/tarmak/utils"
)

const (
	timeLayout = "Jan _2 15:04:05"
)

type Logs struct {
	tarmak       interfaces.Tarmak
	ssh          interfaces.SSH
	instancePool interfaces.InstancePool
	ctx          interfaces.CancellationContext
	log          *logrus.Entry

	path  string
	since string
	until string

	mu sync.Mutex
	wg sync.WaitGroup
}

type SystemdEntry struct {
	Cursor                  string `json:"__CURSOR"`
	RealtimeTimestamp       int64  `json:"__REALTIME_TIMESTAMP,string"`
	MonotonicTimestamp      string `json:"__MONOTONIC_TIMESTAMP"`
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
	host         string
	journaldLogs map[string][]*SystemdEntry
}

func New(tarmak interfaces.Tarmak) *Logs {
	return &Logs{
		tarmak: tarmak,
		log:    tarmak.Log(),
		ctx:    tarmak.CancellationContext(),
	}
}

func (l *Logs) Gather(pool string, flags tarmakv1alpha1.ClusterLogsFlags) error {
	l.log.Infof("fetching logs from target '%s'", pool)

	if i := l.tarmak.Cluster().InstancePool(pool); i == nil {
		return fmt.Errorf("unable to find instance pool '%s'", pool)
	}

	if err := l.initialise(pool, flags); err != nil {
		return fmt.Errorf("failed to set tar gz log bundle path: %v", err)
	}

	err := l.ssh.WriteConfig(l.tarmak.Cluster())
	if err != nil {
		return err
	}

	hosts, err := l.tarmak.Cluster().ListHosts()
	if err != nil {
		return err
	}

	select {
	case <-l.ctx.Done():
		return l.ctx.Err()
	default:
	}

	var poolLogs []*instanceLogs
	var result *multierror.Error

	hostGatherFunc := func(host string) {
		defer l.wg.Done()

		l.log.Infof("fetching journald logs from instance '%s'", host)
		raw, err := l.fetchCmdOutput(
			host,
			"journalctl",
			[]string{"-o",
				"json",
				"--no-pager",
				"--since",
				l.since,
				"--until",
				l.until,
			},
		)
		if err != nil {
			l.mu.Lock()
			result = multierror.Append(result, err)
			l.mu.Unlock()
			return
		}

		journaldLogs := make(map[string][]*SystemdEntry)
		for _, r := range bytes.Split(raw, []byte("\n")) {
			if r == nil || len(r) == 0 {
				continue
			}

			entry := new(SystemdEntry)
			if err := json.Unmarshal(r, entry); err != nil {
				l.mu.Lock()
				err = fmt.Errorf("failed to unmarshal entry [%s]: %s", r, err)
				result = multierror.Append(result, err)
				l.mu.Unlock()
				return
			}

			journaldLogs[entry.SyslogIdentifier] = append(journaldLogs[entry.SyslogIdentifier], entry)

			select {
			case <-l.ctx.Done():
				return
			default:
			}
		}

		l.mu.Lock()
		poolLogs = append(poolLogs, &instanceLogs{host, journaldLogs})
		l.mu.Unlock()
	}

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

		l.wg.Add(1)
		go hostGatherFunc(name)
	}

	l.wg.Wait()

	select {
	case <-l.ctx.Done():
		return l.ctx.Err()
	default:
	}

	if result != nil {
		return result.ErrorOrNil()
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

		for unit, entries := range i.journaldLogs {
			var fileData []byte

			for _, entry := range entries {
				t := time.Unix(entry.RealtimeTimestamp/1000000, 0)
				fileData = append(
					fileData,
					[]byte(fmt.Sprintf("%s %s %s[%s]: %v\n",
						t.Format(timeLayout),
						entry.Hostname,
						entry.SyslogIdentifier,
						entry.Pid,
						entry.Message),
					)...,
				)
			}

			ufile := filepath.Join(idir, fmt.Sprintf(
				"%s.log",
				strings.Replace(unit, "/", "-", -1)),
			)
			f, err := os.Create(ufile)
			if err != nil {
				result = multierror.Append(result, err)
				continue
			}

			if _, err := f.Write(fileData); err != nil {
				result = multierror.Append(result, err)
			}
			f.Close()
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
		cmdStr := fmt.Sprintf("%s %s", command, strings.Join(args, " "))
		return nil, fmt.Errorf("command [%s] returned non-zero: %d", cmdStr, ret)
	}

	return stdout.Bytes(), err
}

func (l *Logs) initialise(pool string, flags tarmakv1alpha1.ClusterLogsFlags) error {
	if flags.Path == utils.DefaultLogsPathPlaceholder {
		l.path = filepath.Join(l.tarmak.Cluster().ConfigPath(), fmt.Sprintf("%s-logs.tar.gz", pool))
	} else {
		p, err := homedir.Expand(flags.Path)
		if err != nil {
			return err
		}

		p, err = filepath.Abs(p)
		if err != nil {
			return err
		}
		l.path = p
	}

	l.ssh = l.tarmak.SSH()
	l.since = fmt.Sprintf(`"%s"`, flags.Since)
	l.until = fmt.Sprintf(`"%s"`, flags.Until)

	return nil
}
