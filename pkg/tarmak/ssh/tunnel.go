package ssh

import (
	"fmt"
	"io"
	"net"
	"os/exec"
	"syscall"
	"time"

	"github.com/Sirupsen/logrus"

	"github.com/jetstack/tarmak/pkg/tarmak/interfaces"
	"github.com/jetstack/tarmak/pkg/tarmak/utils"
)

type Tunnel struct {
	cmd       *exec.Cmd
	running   chan struct{}
	localPort int
	log       *logrus.Entry
	stdin     io.WriteCloser

	retryCount int
	retryWait  time.Duration
}

var _ interfaces.Tunnel = &Tunnel{}

// This opens a local tunnel through a SSH connection
func (s *SSH) Tunnel(hostname string, destination string, destinationPort int) interfaces.Tunnel {
	t := &Tunnel{
		localPort:  utils.UnusedPort(),
		running:    make(chan struct{}, 0),
		log:        s.log.WithField("destination", destination),
		retryCount: 30,
		retryWait:  500 * time.Millisecond,
	}

	args := append(s.args(), "-N", fmt.Sprintf("-L%s:%d:%s:%d", t.BindAddress(), t.localPort, destination, destinationPort), "bastion")

	t.cmd = exec.Command(args[0], args[1:len(args)]...)

	return t
}

// Start tunnel and wait till a tcp socket is reachable
func (t *Tunnel) Start() error {
	var err error

	t.stdin, err = t.cmd.StdinPipe()
	if err != nil {
		return err
	}

	t.log.Debugf("start tunnel cmd=%s", t.cmd.Args)

	err = t.cmd.Start()
	if err != nil {
		return err
	}

	// watch for a terminated SSH
	go func() {
		err := t.cmd.Wait()
		if err != nil {
			perr, ok := err.(*exec.ExitError)
			if ok {
				if status, ok := perr.Sys().(syscall.WaitStatus); ok && status.ExitStatus() == 255 {
					err = nil
				}
			}
		}

		if err != nil {
			t.log.Warn("ssh tunnel stopped with error: ", err)
		} else {
			t.log.Debug("tunnel stopped")
		}
		close(t.running)
	}()

	// wait for TCP socket to be reachable
	tries := t.retryCount
	for {
		select {
		case _, open := <-t.running:
			if !open {
				return fmt.Errorf("ssh is no longer running")
			}
		default:
		}

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
	if err := t.cmd.Process.Signal(syscall.SIGTERM); err != nil {
		t.log.Warnf("error sending signal to process failed: %s", err)
	}
	t.stdin.Close()

	<-t.running

	return nil
}

func (t *Tunnel) Port() int {
	return t.localPort
}

func (t *Tunnel) BindAddress() string {
	return "localhost"
}
