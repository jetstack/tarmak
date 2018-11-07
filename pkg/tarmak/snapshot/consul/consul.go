// Copyright Jetstack Ltd. See LICENSE for details.
package consul

import (
	"bufio"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/sirupsen/logrus"

	clusterv1alpha1 "github.com/jetstack/tarmak/pkg/apis/cluster/v1alpha1"
	"github.com/jetstack/tarmak/pkg/tarmak/interfaces"
	"github.com/jetstack/tarmak/pkg/tarmak/snapshot"
)

var _ interfaces.Snapshot = &Consul{}

const (
	snapshotTimeLayout = "2006-01-02_15-04-05"
)

var (
	exportCmd = []string{
		"export",
		"CONSUL_HTTP_TOKEN=$(sudo cat /etc/consul/consul.json | jq -r '.acl_master_token');",
	}
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
	targetPath := fmt.Sprintf("/tmp/consul-snapshot-%s.snap", time.Now().Format(snapshotTimeLayout))

	cmdArgs := append(exportCmd, "consul", "snapshot", "save", targetPath)
	err = c.sshCmd(
		aliases[0],
		cmdArgs[0],
		cmdArgs[1:],
	)
	if err != nil {
		return err
	}

	ret, err := c.ssh.ScpToLocal(aliases[0], targetPath, c.path)
	if ret != 0 {
		cmdStr := fmt.Sprintf("%s", strings.Join(cmdArgs, " "))
		return fmt.Errorf("command [%s] returned non-zero: %d, %s", cmdStr, ret, err)
	}

	c.log.Infof("consul snapshot saved to %s", c.path)

	return err
}

func (c *Consul) Restore() error {
	aliases, err := snapshot.Prepare(c.tarmak, clusterv1alpha1.InstancePoolTypeVault)
	if err != nil {
		return err
	}
	c.aliases = aliases

	for _, a := range aliases {
		c.log.Infof("restoring snapshot to instance %s", a)
		targetPath := fmt.Sprintf("/tmp/consul-snapshot-%s.snap", time.Now().Format(snapshotTimeLayout))

		ret, err := c.ssh.ScpToHost(a, c.path, targetPath)
		if ret != 0 {
			return fmt.Errorf("command scp returned non-zero: %d, %s", ret, err)
		}

		cmdArgs := append(exportCmd, "consul", "snapshot", "restore", targetPath)
		err = c.sshCmd(
			a,
			cmdArgs[0],
			cmdArgs[1:],
		)
		if err != nil {
			return err
		}
	}

	c.log.Infof("consul snapshot restored from %s", c.path)

	return err
}

func (c *Consul) sshCmd(host, command string, args []string) error {
	readerO, writerO := io.Pipe()
	readerE, writerE := io.Pipe()
	scannerO := bufio.NewScanner(readerO)
	scannerE := bufio.NewScanner(readerE)

	go func() {
		for scannerO.Scan() {
			c.log.WithField("std", "out").Debug(scannerO.Text())
		}
	}()

	go func() {
		for scannerE.Scan() {
			c.log.WithField("std", "err").Warn(scannerE.Text())
		}
	}()

	ret, err := c.ssh.ExecuteWithWriter(host, command, args, writerO, writerE)
	if ret != 0 {
		cmdStr := fmt.Sprintf("%s %s", command, strings.Join(args, " "))
		return fmt.Errorf("command [%s] returned non-zero: %d", cmdStr, ret)
	}

	return err
}
