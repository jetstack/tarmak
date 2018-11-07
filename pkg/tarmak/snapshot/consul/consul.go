// Copyright Jetstack Ltd. See LICENSE for details.
package consul

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/hashicorp/go-multierror"
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

	var result *multierror.Error
	var errLock sync.Mutex

	reader, writer := io.Pipe()
	go snapshot.ReadTarFromStream(c.path, reader, result, errLock)

	hostPath := fmt.Sprintf("/tmp/consul-snapshot-%s.snap",
		time.Now().Format(snapshot.TimeLayout))
	cmdArgs := append(envCmd,
		strings.Split(fmt.Sprintf(consulCmd, "save", hostPath), " ")...)
	cmdArgs = append(cmdArgs,
		strings.Split(fmt.Sprintf(snapshot.TarCCmd, hostPath), " ")...)

	err = c.sshCmd(
		aliases[0],
		cmdArgs,
		nil,
		writer,
	)
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

		hostPath := fmt.Sprintf("/tmp/consul-snapshot-%s.snap",
			time.Now().Format(snapshot.TimeLayout))

		cmdArgs := strings.Split(fmt.Sprintf(snapshot.TarXCmd, hostPath), " ")
		//cmdArgs := strings.Split(snapshot.TarXCmd, " ")
		//cmdArgs = append(cmdArgs,
		//	append(envCmd,
		//		strings.Split(fmt.Sprintf(consulCmd, "restore", hostPath), " ")...)...)

		var result *multierror.Error
		var errLock sync.Mutex

		reader, writer := io.Pipe()
		go snapshot.WriteTarToStream(c.path, writer, result, errLock)

		err = c.sshCmd(
			a,
			cmdArgs,
			reader,
			os.Stdout,
		)
		if err != nil {
			return err
		}
	}

	c.log.Infof("consul snapshot restored from %s", c.path)

	return nil
}

func (c *Consul) sshCmd(host string, args []string, stdin io.Reader, stdout io.Writer) error {
	readerE, writerE := io.Pipe()
	scannerE := bufio.NewScanner(readerE)

	go func() {
		for scannerE.Scan() {
			c.log.WithField("std", "err").Warn(scannerE.Text())
		}
	}()

	ret, err := c.ssh.ExecuteWithPipe(host, args[0], args[1:], stdin, stdout, writerE)
	if ret != 0 {
		return fmt.Errorf("command [%s] returned non-zero: %d", strings.Join(args, " "), ret)
	}

	return err
}
