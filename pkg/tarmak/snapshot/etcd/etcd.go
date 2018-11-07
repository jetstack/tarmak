// Copyright Jetstack Ltd. See LICENSE for details.
package etcd

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	//"github.com/docker/docker/pkg/archive"
	"github.com/hashicorp/go-multierror"
	"github.com/sirupsen/logrus"

	clusterv1alpha1 "github.com/jetstack/tarmak/pkg/apis/cluster/v1alpha1"
	"github.com/jetstack/tarmak/pkg/tarmak/interfaces"
	"github.com/jetstack/tarmak/pkg/tarmak/snapshot"
)

var _ interfaces.Snapshot = &Etcd{}

const (
	snapshotTimeLayout = "2006-01-02_15-04-05"

	etcdctlCmd = "/opt/bin/etcdctl snapshot %s %s > /dev/null;"
	tarCmd     = "tar -czPf - %s"
)

var (
	stores = []map[string]string{
		{"store": "k8s-main", "file": "k8s", "port": "2379"},
		{"store": "k8s-events", "file": "k8s", "port": "2369"},
		{"store": "overlay", "file": "overlay", "port": "2359"},
	}

	exportCmd = []string{
		"ETCDCTL_CERT=/etc/etcd/ssl/etcd-{{file}}.pem",
		"ETCDCTL_KEY=/etc/etcd/ssl/etcd-{{file}}-key.pem",
		"ETCDCTL_CACERT=/etc/etcd/ssl/etcd-{{file}}-ca.pem",
		"ETCDCTL_API=3",
		"ETCDCTL_ENDPOINTS=https://127.0.0.1:{{port}}",
	}
)

type Etcd struct {
	tarmak interfaces.Tarmak
	ssh    interfaces.SSH
	log    *logrus.Entry
	ctx    interfaces.CancellationContext

	path    string
	aliases []string
	errLock sync.Mutex // prevent multiple writes to result error
}

func New(tarmak interfaces.Tarmak, path string) *Etcd {
	return &Etcd{
		tarmak: tarmak,
		ssh:    tarmak.SSH(),
		ctx:    tarmak.CancellationContext(),
		log:    tarmak.Log(),
		path:   path,
	}
}

func (e *Etcd) Save() error {
	aliases, err := snapshot.Prepare(e.tarmak, clusterv1alpha1.InstancePoolTypeEtcd)
	if err != nil {
		return err
	}
	e.aliases = aliases

	e.log.Infof("saving snapshots from instance %s", aliases[0])

	var wg sync.WaitGroup
	var result *multierror.Error

	saveFunc := func(store map[string]string) {
		defer wg.Done()

		hostPath := fmt.Sprintf("/tmp/etcd-snapshot-%s-%s.db",
			store["store"], time.Now().Format(snapshotTimeLayout))
		targetPath := fmt.Sprintf("%s%s.db", e.path, store["store"])

		cmdArgs := append(e.template(exportCmd, store),
			strings.Split(fmt.Sprintf(etcdctlCmd, "save", hostPath), " ")...)
		cmdArgs = append(cmdArgs,
			strings.Split(fmt.Sprintf(tarCmd, hostPath), " ")...)

		reader, writer := io.Pipe()
		go e.readTarFromStream(targetPath, reader, result)

		err = e.sshCmd(
			aliases[0],
			cmdArgs,
			writer,
		)
		if err != nil {

			e.errLock.Lock()
			result = multierror.Append(result, err)
			e.errLock.Unlock()

			return
		}

		e.log.Infof("etcd %s snapshot saved to %s", store["store"], targetPath)

		select {
		case <-e.ctx.Done():
			return
		default:
		}
	}

	wg.Add(len(stores))

	for _, store := range stores {
		go saveFunc(store)
	}

	wg.Wait()

	select {
	case <-e.ctx.Done():
		return e.ctx.Err()
	default:
	}

	return result.ErrorOrNil()
}

func (e *Etcd) Restore() error {
	aliases, err := snapshot.Prepare(e.tarmak, clusterv1alpha1.InstancePoolTypeEtcd)
	if err != nil {
		return err
	}
	e.aliases = aliases

	return nil
}

func (e *Etcd) sshCmd(host string, args []string, stdout io.Writer) error {
	readerE, writerE := io.Pipe()
	scannerE := bufio.NewScanner(readerE)

	go func() {
		for scannerE.Scan() {
			e.log.WithField("std", "err").Warn(scannerE.Text())
		}
	}()

	args = append([]string{"sudo"}, args...)
	ret, err := e.ssh.ExecuteWithWriter(host, args[0], args[1:], stdout, writerE)
	if ret != 0 {
		cmdStr := fmt.Sprintf("%s", strings.Join(args, " "))
		return fmt.Errorf("command [%s] returned non-zero: %d", cmdStr, ret)
	}

	return err
}

func (e *Etcd) readTarFromStream(dest string, stream io.Reader, result *multierror.Error) {
	gzr, err := gzip.NewReader(stream)
	if err != nil {

		e.errLock.Lock()
		result = multierror.Append(result, err)
		e.errLock.Unlock()

		return
	}

	f, err := os.Create(dest)
	if err != nil {

		e.errLock.Lock()
		result = multierror.Append(result, err)
		e.errLock.Unlock()

		return
	}

	if _, err := io.Copy(f, gzr); err != nil {

		e.errLock.Lock()
		result = multierror.Append(result, err)
		e.errLock.Unlock()

		return
	}
}

func (e *Etcd) template(args []string, vars map[string]string) []string {
	for i := range args {
		for k, v := range vars {
			args[i] = strings.Replace(args[i], fmt.Sprintf("{{%s}}", k), v, -1)
		}
	}

	return args
}
