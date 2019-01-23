// Copyright Jetstack Ltd. See LICENSE for details.
package etcd

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	multierror "github.com/hashicorp/go-multierror"
	"github.com/sirupsen/logrus"

	clusterv1alpha1 "github.com/jetstack/tarmak/pkg/apis/cluster/v1alpha1"
	"github.com/jetstack/tarmak/pkg/tarmak/interfaces"
	"github.com/jetstack/tarmak/pkg/tarmak/snapshot"
	"github.com/jetstack/tarmak/pkg/tarmak/utils/consts"
)

var _ interfaces.Snapshot = &Etcd{}

const (
	etcdctlCmd = `/opt/bin/etcdctl snapshot %s %s;`
)

var (
	stores = []map[string]string{
		{"cluster": consts.RestoreK8sMainFlagName, "file": "k8s", "client_port": "2379", "peer_port": "2380"},
		{"cluster": consts.RestoreK8sEventsFlagName, "file": "k8s", "client_port": "2369", "peer_port": "2370"},
		{"cluster": consts.RestoreOverlayFlagName, "file": "overlay", "client_port": "2359", "peer_port": "2360"},
	}

	envCmd = `
set -e;
export ETCDCTL_CERT=/etc/etcd/ssl/etcd-{{file}}.pem;
export ETCDCTL_KEY=/etc/etcd/ssl/etcd-{{file}}-key.pem;
export ETCDCTL_CACERT=/etc/etcd/ssl/etcd-{{file}}-ca.pem;
export ETCDCTL_API=3;
export ETCDCTL_ENDPOINTS=https://127.0.0.1:{{client_port}};
`
)

type Etcd struct {
	tarmak interfaces.Tarmak
	ssh    interfaces.SSH
	log    *logrus.Entry
	ctx    interfaces.CancellationContext

	path string
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

	e.log.Infof("saving snapshots from instance %s", aliases[0])

	var wg sync.WaitGroup
	var result *multierror.Error
	var errLock sync.Mutex

	saveFunc := func(store map[string]string) {
		defer wg.Done()

		hostPath := fmt.Sprintf("/tmp/etcd-snapshot-%s-%s.db",
			store["cluster"], time.Now().Format(snapshot.TimeLayout))

		cmdArgs := fmt.Sprintf(
			`sudo /bin/bash -c "%s %s %s"`,
			e.template(envCmd, store),
			fmt.Sprintf(etcdctlCmd, "save", hostPath),
			fmt.Sprintf(etcdctlCmd, "status", hostPath),
		)

		err = snapshot.SSHCmd(e, aliases[0], cmdArgs, nil, nil, nil)
		if err != nil {

			errLock.Lock()
			result = multierror.Append(result, err)
			errLock.Unlock()

			return
		}

		targetPath := fmt.Sprintf("%s%s.db", e.path, store["cluster"])
		reader, writer := io.Pipe()

		err = snapshot.TarFromStream(func() error {
			err := snapshot.SSHCmd(e, aliases[0], fmt.Sprintf(snapshot.GZipCCmd, hostPath),
				nil, writer, nil)
			writer.Close()
			return err

		}, reader, targetPath)

		if err != nil {

			errLock.Lock()
			result = multierror.Append(result, err)
			errLock.Unlock()

			return
		}

		err = snapshot.SSHCmd(e, aliases[0], fmt.Sprintf("sudo rm %s", hostPath), nil, nil, nil)
		if err != nil {

			errLock.Lock()
			result = multierror.Append(result, err)
			errLock.Unlock()

			return
		}

		e.log.Infof("etcd %s snapshot saved to %s", store["cluster"], targetPath)

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

	stopEtcdFunc := func(host string, store map[string]string) error {
		cmdArgs := fmt.Sprintf(`set -e; sudo systemctl stop etcd-%s `, store["cluster"])
		err = snapshot.SSHCmd(e, host, cmdArgs, nil, nil, nil)

		return err
	}

	//restoreFunc := func(host, path, token string, store map[string]string) error {
	//	reader, writer := io.Pipe()
	//	hostPath := fmt.Sprintf("/tmp/etcd-snapshot-%s-%s.db",
	//		store["cluster"], time.Now().Format(snapshot.TimeLayout))

	//	err = snapshot.TarToStream(func() error {
	//		err := snapshot.SSHCmd(e, host, fmt.Sprintf(snapshot.GZipDCmd, hostPath), reader, nil, nil)
	//		return err
	//	}, writer, path)
	//	if err != nil {
	//		return err
	//	}

	//
	//		cmdArgs = e.template(`set -e;
	//sudo mkdir -p /var/lib/etcd_backup;
	//sudo rsync -a --delete --ignore-missing-args /var/lib/etcd/{{cluster}} /var/lib/etcd_backup/;
	//sudo rm -rf /var/lib/etcd/{{cluster}};
	//`, store)
	//		err = snapshot.SSHCmd(e, host, cmdArgs, nil, nil, nil)
	//		if err != nil {
	//			return err
	//		}
	//
	//		initialCluster := e.initialClusterString(host, store)
	//		for _, a := range aliases[1:] {
	//			initialCluster = strings.Join(
	//				[]string{initialCluster, e.initialClusterString(a, store)}, ",",
	//			)
	//		}
	//
	//		cmdArgs = e.template(fmt.Sprintf(`set -e;
	//sudo ETCDCTL_API=3 /opt/bin/etcdctl snapshot restore %s \
	//--name=%s.%s.%s \
	//--data-dir=/var/lib/etcd/{{cluster}} \
	//--initial-advertise-peer-urls=https://%s.%s.%s:{{peer_port}} \
	//--initial-cluster=%s \
	//--initial-cluster-token=etcd-{{cluster}}-%s
	//`,
	//			hostPath,
	//			host, e.clusterName(), e.privateZone(),
	//			host, e.clusterName(), e.privateZone(),
	//			initialCluster,
	//			token,
	//		), store)
	//		err = snapshot.SSHCmd(e, host, cmdArgs, nil, nil, nil)
	//		if err != nil {
	//			return err
	//		}
	//
	//		cmdArgs = e.template(`set -e;
	// sudo chown -R etcd:etcd /var/lib/etcd/{{cluster}}
	// `, store)
	//		err = snapshot.SSHCmd(e, host, cmdArgs, nil, nil, nil)
	//		if err != nil {
	//			return err
	//		}
	//
	//		return nil
	//	}
	//
	startEtcdFunc := func(host string, store map[string]string) error {
		cmdArgs := fmt.Sprintf(`set -e; sudo systemctl start etcd-%s`, store["cluster"])
		err = snapshot.SSHCmd(e, host, cmdArgs, nil, nil, nil)
		if err != nil {
			return err
		}

		return nil
	}

	healthCheckFunc := func(host string, store map[string]string) error {
		endpoints := e.endpointsString(host, store)
		for _, a := range aliases[1:] {
			endpoints = strings.Join(
				[]string{endpoints, e.endpointsString(a, store)}, ",",
			)
		}

		cmdArgs := e.template(fmt.Sprint(`%s; sudo /opt/bin/etcdctl endpoint health --endpoints=%s `,
			envCmd, endpoints), store)
		err = snapshot.SSHCmd(e, host, cmdArgs, nil, nil, nil)
		if err != nil {
			return err
		}

		return nil
	}

	for _, store := range stores {
		value := e.restoreFlagValue(store["cluster"])
		if value == "" {
			continue
		}

		b := make([]byte, 32)
		_, err := rand.Read(b)
		if err != nil {
			return fmt.Errorf("failed to create random etcd initial token: %s", err)
		}
		token := base64.URLEncoding.EncodeToString(b)

		fmt.Printf(">%s\n", token)

		for _, a := range aliases {
			e.log.Infof("stopping etcd unit %s on host %s", store["cluster"], a)
			err := stopEtcdFunc(a, store)
			if err != nil {
				return err
			}
		}

		time.Sleep(5 * time.Second)

		//for _, a := range aliases {
		//	e.log.Infof("restoring etcd %s on host %s", store["cluster"], a)
		//	err := restoreFunc(a, value, token, store)
		//	if err != nil {
		//		return err
		//	}
		//}

		var wg sync.WaitGroup
		var result *multierror.Error
		var errLock sync.Mutex

		wg.Add(len(aliases))

		for _, a := range aliases {
			e.log.Infof("starting etcd %s on host %s", store["cluster"], a)

			go func() {
				defer wg.Done()

				err := startEtcdFunc(a, store)
				if err != nil {

					errLock.Lock()
					result = multierror.Append(result, err)
					errLock.Unlock()

				}

			}()
		}

		wg.Wait()

		if result != nil {
			return result
		}

		for _, a := range aliases {
			e.log.Infof("checking health of etcd %s on host %s", store["cluster"], a)
			err := healthCheckFunc(a, store)
			if err != nil {
				return err
			}
		}

		e.log.Infof("successfully restored etcd cluster %s with snapshot %s", store["cluster"], value)

		select {
		case <-e.tarmak.CancellationContext().Done():
			return e.tarmak.CancellationContext().Err()
		default:
		}
	}

	e.log.Info("restarting API servers on master hosts")
	//masters, err := snapshot.Prepare(e.tarmak, clusterv1alpha1.InstancePoolTypeMaster)
	//for _, master := range masters {
	//	cmdArgs := " sudo systemctl restart kube-apiserver"
	//	err = snapshot.SSHCmd(e, master, cmdArgs, nil, nil, nil)
	//	if err != nil {
	//		return err
	//	}
	//}

	return nil
}

func (e *Etcd) template(args string, vars map[string]string) string {
	for k, v := range vars {
		args = strings.Replace(args, fmt.Sprintf("{{%s}}", k), v, -1)
	}

	return args
}

func (e *Etcd) restoreFlagValue(flag string) string {
	rf := e.tarmak.ClusterFlags().Snapshot.Etcd.Restore
	for _, db := range []struct {
		name, value string
	}{
		{consts.RestoreK8sMainFlagName, rf.K8sMain},
		{consts.RestoreK8sEventsFlagName, rf.K8sEvents},
		{consts.RestoreOverlayFlagName, rf.Overlay},
	} {
		if db.name == flag {
			return db.value
		}
	}

	return ""
}

func (e *Etcd) endpointsString(host string, store map[string]string) string {
	return fmt.Sprintf("%s.%s.%s=https://%s.%s.%s:%s",
		host, e.clusterName(), e.privateZone(),
		host, e.clusterName(), e.privateZone(), store["client_port"])
}

func (e *Etcd) Log() *logrus.Entry {
	return e.log
}

func (e *Etcd) SSH() interfaces.SSH {
	return e.ssh
}

func (e *Etcd) clusterName() string {
	return e.tarmak.Cluster().ClusterName()
}

func (e *Etcd) privateZone() string {
	return e.tarmak.Environment().Config().PrivateZone
}
