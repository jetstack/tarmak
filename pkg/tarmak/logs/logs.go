// Copyright Jetstack Ltd. See LICENSE for details.
package logs

import (
	"bufio"
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

var (
	TargetGroups = []string{
		"bastion",
		"vault",
		"etcd",
		"worker",
		"control-plane",
	}
)

type Logs struct {
	tarmak interfaces.Tarmak
	ssh    interfaces.SSH
	ctx    interfaces.CancellationContext
	log    *logrus.Entry

	path    string
	since   string
	until   string
	targets []string

	mu       sync.Mutex
	wg       sync.WaitGroup
	hosts    []interfaces.Host
	tmpDir   string
	tmpFiles map[string]*os.File
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

func New(tarmak interfaces.Tarmak) *Logs {
	return &Logs{
		tarmak: tarmak,
		log:    tarmak.Log(),
		ctx:    tarmak.CancellationContext(),
	}
}

func (l *Logs) Gather(group string, flags tarmakv1alpha1.ClusterLogsFlags) error {
	l.log.Infof("fetching logs from target '%s'", group)
	if err := l.initialise(group, flags); err != nil {
		return fmt.Errorf("failed to set tar gz log bundle path: %v", err)
	}

	err := l.ssh.WriteConfig(l.tarmak.Cluster())
	if err != nil {
		return err
	}

	select {
	case <-l.ctx.Done():
		return l.ctx.Err()
	default:
	}

	dir, err := ioutil.TempDir("", filepath.Base(l.path))
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)
	l.tmpDir = dir

	var result *multierror.Error
	hostGatherFunc := func(host string) {
		defer l.wg.Done()

		reader, writer := io.Pipe()
		go l.readStream(reader, host, result)

		l.log.Infof("fetching journald logs from instance '%s'", host)
		err := l.fetchCmdOutput(
			host,
			"journalctl",
			[]string{"-o", "json", "--no-pager", "--since", l.since, "--until", l.until},
			writer,
		)
		if err != nil {
			l.mu.Lock()
			result = multierror.Append(result, err)
			l.mu.Unlock()
		}
	}

	aliases, err := l.hostAliases()
	if err != nil {
		return err
	}

	for _, a := range aliases {
		if err := os.Mkdir(
			filepath.Join(l.tmpDir, a),
			os.FileMode(0755)); err != nil {
			return err
		}

		l.wg.Add(1)
		go hostGatherFunc(a)
	}

	l.wg.Wait()

	if result != nil {
		return result.ErrorOrNil()
	}

	select {
	case <-l.ctx.Done():
		return l.ctx.Err()
	default:
	}

	for _, f := range l.tmpFiles {
		f.Close()
	}

	return l.bundleLogs()
}

func (l *Logs) readStream(reader io.Reader, host string, result *multierror.Error) {
	readerScanner := bufio.NewScanner(reader)

	for readerScanner.Scan() {
		r := readerScanner.Bytes()

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

		if err := l.writeToFile(host, entry); err != nil {
			l.mu.Lock()
			result = multierror.Append(result, err)
			l.mu.Unlock()
			return
		}

		select {
		case <-l.ctx.Done():
			return
		default:
		}
	}
}

// return all aliases of instances in the target group
func (l *Logs) hostAliases() ([]string, error) {
	var aliases []string
	var result *multierror.Error

	for _, host := range l.hosts {
		for _, target := range l.targets {
			if utils.SliceContainsPrefix(host.Roles(), target) {
				if len(host.Aliases()) == 0 {
					err := fmt.Errorf(
						"host with correct role '%v' found without alias: %v",
						host.Roles(),
						host.ID(),
					)
					result = multierror.Append(result, err)
					break
				}

				aliases = append(aliases, host.Aliases()[0])
				break
			}
		}
	}

	return aliases, result.ErrorOrNil()
}

func (l *Logs) writeToFile(host string, entry *SystemdEntry) error {
	fname := filepath.Join(l.tmpDir, host,
		fmt.Sprintf("%s.log",
			strings.Replace(entry.SyslogIdentifier, "/", "-", -1),
		),
	)

	f, ok := l.tmpFiles[fname]
	if !ok || f == nil {
		var err error
		f, err = os.Create(fname)
		if err != nil {
			return err
		}
		l.tmpFiles[fname] = f
	}

	t := time.Unix(entry.RealtimeTimestamp/1000000, 0)
	_, err := f.Write([]byte(fmt.Sprintf("%s %s %s[%s]: %v\n",
		t.Format(timeLayout),
		entry.Hostname,
		entry.SyslogIdentifier,
		entry.Pid,
		entry.Message),
	))

	return err
}

func (l *Logs) bundleLogs() error {
	f, err := os.Create(l.path)
	if err != nil {
		return err
	}
	defer f.Close()

	reader, err := archive.Tar(
		l.tmpDir,
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

func (l *Logs) fetchCmdOutput(host, command string, args []string, stdout io.Writer) error {
	ret, err := l.ssh.ExecuteWithWriter(host, command, args, stdout)
	if ret != 0 {
		cmdStr := fmt.Sprintf("%s %s", command, strings.Join(args, " "))
		return fmt.Errorf("command [%s] returned non-zero: %d", cmdStr, ret)
	}

	return err
}

func (l *Logs) initialise(group string, flags tarmakv1alpha1.ClusterLogsFlags) error {
	if flags.Path == utils.DefaultLogsPathPlaceholder {
		l.path = filepath.Join(l.tarmak.Cluster().ConfigPath(), fmt.Sprintf("%s-logs.tar.gz", group))
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

	// only need to list hosts once so we can cache the list
	if len(l.hosts) == 0 {
		hosts, err := l.tarmak.Cluster().ListHosts()
		if err != nil {
			return err
		}
		l.hosts = hosts
	}

	l.ssh = l.tarmak.SSH()
	l.since = fmt.Sprintf(`"%s"`, flags.Since)
	l.until = fmt.Sprintf(`"%s"`, flags.Until)
	l.tmpFiles = make(map[string]*os.File)

	switch group {
	case "control-plane":
		l.targets = []string{"master", "etcd"}
	default:
		l.targets = []string{group}
	}

	return nil
}
