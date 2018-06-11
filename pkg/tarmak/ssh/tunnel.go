// Copyright Jetstack Ltd. See LICENSE for details.
package ssh

import (
	"fmt"
	"io"
	"net"
	"os/exec"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/jetstack/tarmak/pkg/tarmak/interfaces"
	"github.com/jetstack/tarmak/pkg/tarmak/utils"
)

type Tunnel struct {
	localPort int
	log       *logrus.Entry
	stdin     io.WriteCloser

	retryCount int
	retryWait  time.Duration

	forwardSpec string
	sshCommand  []string
}

var _ interfaces.Tunnel = &Tunnel{}

// This opens a local tunnel through a SSH connection
func (s *SSH) Tunnel(hostname string, destination string, destinationPort int) interfaces.Tunnel {
	t := &Tunnel{
		localPort:  utils.UnusedPort(),
		log:        s.log.WithField("destination", destination),
		retryCount: 30,
		retryWait:  500 * time.Millisecond,
		sshCommand: s.args(),
	}
	t.forwardSpec = fmt.Sprintf("-L%s:%d:%s:%d", t.BindAddress(), t.localPort, destination, destinationPort)

	return t
}

// Start tunnel and wait till a tcp socket is reachable
func (t *Tunnel) Start() error {
	var err error

	// ensure there is connectivity to the bastion
	args := append(t.sshCommand, "bastion", "/bin/true")
	cmd := exec.Command(args[0], args[1:len(args)]...)

	t.log.Debugf("check SSH connection to bastion cmd=%s", cmd.Args)
	err = cmd.Start()
	if err != nil {
		return err
	}

	// check for errors
	err = cmd.Wait()
	if err != nil {
		return fmt.Errorf("error checking SSH connecting to bastion: %s", err)
	}

	args = append(t.sshCommand, "-O", "forward", t.forwardSpec, "bastion")
	cmd = exec.Command(args[0], args[1:len(args)]...)

	t.log.Debugf("start tunnel cmd=%s", cmd.Args)
	err = cmd.Start()
	if err != nil {
		return err
	}

	// check for errors
	err = cmd.Wait()
	if err != nil {
		return fmt.Errorf("error starting SSH tunnel via bastion: %s", err)
	}

	// wait for TCP socket to be reachable
	tries := t.retryCount
	for {
		if conn, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", t.Port()), t.retryWait); err != nil {
			t.log.Debug("error connecting to tunnel: ", err)
		} else {
			conn.Close()
			return nil
		}

		tries -= 1
		if tries == 0 {
			break
		}
		time.Sleep(t.retryWait)
	}

	return fmt.Errorf("could not establish a connection to destination via tunnel after %d tries", t.retryCount)
}

func (t *Tunnel) Stop() error {
	args := append(t.sshCommand, "-O", "cancel", t.forwardSpec, "bastion")
	cmd := exec.Command(args[0], args[1:len(args)]...)

	t.log.Debugf("stop tunnel cmd=%s", cmd.Args)
	err := cmd.Start()
	if err != nil {
		return err
	}

	// check for errors
	err = cmd.Wait()
	if err != nil {
		t.log.Warn("stopping ssh tunnel failed with error: ", err)
	}

	return nil
}

func (t *Tunnel) Port() int {
	return t.localPort
}

func (t *Tunnel) BindAddress() string {
	return "127.0.0.1"
}
