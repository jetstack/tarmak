// Copyright Jetstack Ltd. See LICENSE for details.
package consul

import (
	"fmt"
	"io"
	"time"

	"github.com/sirupsen/logrus"

	clusterv1alpha1 "github.com/jetstack/tarmak/pkg/apis/cluster/v1alpha1"
	"github.com/jetstack/tarmak/pkg/tarmak/interfaces"
	"github.com/jetstack/tarmak/pkg/tarmak/snapshot"
)

var _ interfaces.Snapshot = &Consul{}

const (
	consulCmd = "consul snapshot %s %s > /dev/null;"
)

var (
	envCmd = []string{"CONSUL_HTTP_TOKEN=$(sudo cat /etc/consul/consul.json | jq -r '.acl_master_token')"}
)

type Consul struct {
	tarmak interfaces.Tarmak
	ssh    interfaces.SSH
	log    *logrus.Entry

	path    string
	aliases []string
}

func New(tarmak interfaces.Tarmak, path string) *Consul {
	return &Consul{
		tarmak: tarmak,
		ssh:    tarmak.SSH(),
		log:    tarmak.Log(),
		path:   path,
	}
}

func (c *Consul) Save() error {
	aliases, err := snapshot.Prepare(c.tarmak, clusterv1alpha1.InstancePoolTypeVault)
	if err != nil {
		return err
	}
	c.aliases = aliases

	c.log.Infof("saving snapshot from instance %s", aliases[0])

	hostPath := fmt.Sprintf("/tmp/consul-snapshot-%s.snap",
		time.Now().Format(snapshot.TimeLayout))
	cmdArgs := fmt.Sprintf(`set -e;
export CONSUL_HTTP_TOKEN=$(sudo cat /etc/consul/consul.json | jq -r '.acl_master_token');
export DATACENTER=$(sudo cat /etc/consul/consul.json | jq -r '.datacenter');
/usr/local/bin/consul snapshot save -datacenter $DATACENTER %s;
/usr/local/bin/consul snapshot inspect %s`, hostPath, hostPath)

	err = snapshot.SSHCmd(c, aliases[0], cmdArgs, nil, nil, nil)
	if err != nil {
		return err
	}

	reader, writer := io.Pipe()
	err = snapshot.TarFromStream(func() error {
		err := snapshot.SSHCmd(c, aliases[0], fmt.Sprintf(snapshot.GZipCCmd, hostPath),
			nil, writer, nil)
		writer.Close()
		return err
	}, reader, c.path)
	if err != nil {
		return err
	}

	c.log.Infof("consul snapshot saved to %s", c.path)

	return nil
}

func (c *Consul) Restore() error {
	aliases, err := snapshot.Prepare(c.tarmak, clusterv1alpha1.InstancePoolTypeVault)
	if err != nil {
		return err
	}
	c.aliases = aliases

	for _, a := range aliases {
		c.log.Infof("restoring snapshot to instance %s", a)

		reader, writer := io.Pipe()
		hostPath := fmt.Sprintf("/tmp/consul-snapshot-%s.snap",
			time.Now().Format(snapshot.TimeLayout))

		err = snapshot.TarToStream(func() error {
			err := snapshot.SSHCmd(c, a, fmt.Sprintf(snapshot.GZipDCmd, hostPath), reader, nil, nil)
			return err
		}, writer, c.path)
		if err != nil {
			return err
		}

		cmdArgs := fmt.Sprintf(`set -e;
export CONSUL_HTTP_TOKEN=$(sudo cat /etc/consul/consul.json | jq -r '.acl_master_token');
export DATACENTER=$(sudo cat /etc/consul/consul.json | jq -r '.datacenter');
/usr/local/bin/consul snapshot inspect %s;
/usr/local/bin/consul snapshot restore -datacenter $DATACENTER %s;
echo number of keys: $(curl  --header "X-Consul-Token: $CONSUL_HTTP_TOKEN" -s 'http://127.0.0.1:8500/v1/kv/?keys' | jq '. | length');
`, hostPath, hostPath)

		err = snapshot.SSHCmd(c, a, cmdArgs, nil, nil, nil)
		if err != nil {
			return err
		}
	}

	c.log.Infof("consul snapshot restored from %s", c.path)

	return nil
}

func (c *Consul) Log() *logrus.Entry {
	return c.log
}

func (c *Consul) SSH() interfaces.SSH {
	return c.ssh
}
