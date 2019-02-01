// Copyright Jetstack Ltd. See LICENSE for details.
package logs

import (
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
	logTimeLayout = "2006-01-02 15:04:05" // journalctl time format
	timeLayout    = "Jan _2 15:04:05"     // journald style layout
)

var (
	// target groups of one or more instance groups
	TargetGroups = []string{
		"bastion",
		"vault",
		"etcd",
		"worker",
		"master",
		"control-plane",
	}
)

type Logs struct {
	tarmak interfaces.Tarmak
	ssh    interfaces.SSH
	ctx    interfaces.CancellationContext
	log    *logrus.Entry

	path    string   // target tar ball path
	since   string   // gather logs since datetime
	until   string   // gather logs since datetime
	targets []string // target instance groups

	hosts    []interfaces.Host   //target hosts
	tmpDir   string              // tmp logs dir
	tmpFiles map[string]*os.File // tmp log files

	fileLock  sync.RWMutex   // prevent collisions writing files
	errorLock sync.Mutex     // prevent collisions writing global errors
	wg        sync.WaitGroup // go routine waiter
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
	Unit                    string `json:"UNIT"`
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

func (l *Logs) Aggregate(group string, flags tarmakv1alpha1.ClusterLogsFlags) error {
	l.log.Infof("fetching logs from target '%s'", group)

	if err := l.initialise(group, flags); err != nil {
		return fmt.Errorf("failed to set tar gz log bundle path: %v", err)
	}

	err := l.ssh.WriteConfig(l.tarmak.Cluster())
	if err != nil {
		return err
	}

	aliases, err := l.hostAliases()
	if err != nil {
		return err
	}

	if len(aliases) == 0 {
		return fmt.Errorf("no host aliases found in target '%s'", group)
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
			[]string{"journalctl", "-o", "json", "--no-pager",
				"--since", l.since, "--until", l.until},
			writer,
		)
		if err != nil {

			l.errorLock.Lock()
			result = multierror.Append(result, err)
			l.errorLock.Unlock()

		}
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
		return result
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
	dec := json.NewDecoder(reader)

	for dec.More() {
		entry := new(SystemdEntry)
		if err := dec.Decode(entry); err != nil {

			l.errorLock.Lock()
			result = multierror.Append(result, err)
			l.errorLock.Unlock()

			return
		}

		if err := l.writeToFile(host, entry); err != nil {

			l.errorLock.Lock()
			result = multierror.Append(result, err)
			l.errorLock.Unlock()

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
	unit := entry.SystemdUnit
	if unit == "" {
		unit = entry.Unit
	}
	unit = strings.TrimSuffix(unit, ".service")

	// temporary logs filename
	// some units contain '/' which must be replaced by '-' to save to file
	fname := filepath.Join(l.tmpDir, host,
		fmt.Sprintf("%s.log",
			strings.Replace(unit, "/", "-", -1),
		),
	)

	// get file or create if not existing
	l.fileLock.RLock()
	f, ok := l.tmpFiles[fname]
	l.fileLock.RUnlock()

	if !ok || f == nil {
		var err error
		f, err = os.Create(fname)
		if err != nil {
			return err
		}

		l.fileLock.Lock()
		l.tmpFiles[fname] = f
		l.fileLock.Unlock()

	}

	// expected journalctl formatting
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

func (l *Logs) fetchCmdOutput(host string, cmd []string, stdout io.Writer) error {
	ret, err := l.ssh.Execute(host, cmd, nil, stdout, nil)
	if ret != 0 {
		return fmt.Errorf("command [%s] returned non-zero: %d",
			strings.Join(cmd, " "), ret)
	}

	return err
}

func (l *Logs) initialise(group string, flags tarmakv1alpha1.ClusterLogsFlags) error {
	if flags.Path == utils.DefaultLogsPathPlaceholder {
		wd, err := os.Getwd()
		if err != nil {
			return err
		}

		l.path = filepath.Join(wd, fmt.Sprintf("%s-logs.tar.gz", group))
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
	l.tmpFiles = make(map[string]*os.File)

	now := time.Now()
	if flags.Since == utils.DefaultLogsSincePlaceholder {
		l.since = fmt.Sprintf("'%s'", now.Add(-time.Hour*24).Format(logTimeLayout))
	} else {
		l.since = fmt.Sprintf("'%s'", flags.Since)
	}

	if flags.Until == utils.DefaultLogsUntilPlaceholder {
		l.until = fmt.Sprintf("'%s'", now.Format(logTimeLayout))
	} else {
		l.until = fmt.Sprintf("'%s'", flags.Until)
	}

	switch group {
	case "control-plane":
		l.targets = []string{"vault", "etcd", "master"}
	default:
		l.targets = []string{group}
	}

	return nil
}
