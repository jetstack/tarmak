// Copyright Jetstack Ltd. See LICENSE for details.
package ssh

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"time"

	"github.com/kardianos/osext"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"

	"github.com/jetstack/tarmak/pkg/tarmak/interfaces"
)

type Tunnel struct {
	log    *logrus.Entry
	ssh    *SSH
	stopCh chan struct{}

	dest      string
	destPort  string
	localPort string
	daemonize bool

	bastionConn *ssh.Client
	listener    net.Listener
	daemon      *os.Process

	closeConnsLock sync.Mutex // prevent closing the same connection multiple times at once
	openedConns    []net.Conn
}

var _ interfaces.Tunnel = &Tunnel{}

// This opens a local tunnel through a SSH connection
func (s *SSH) Tunnel(dest, destPort, localPort string, daemonize bool) interfaces.Tunnel {
	tunnel := &Tunnel{
		log:       s.log.WithField("destination", dest),
		ssh:       s,
		dest:      dest,
		destPort:  destPort,
		daemonize: daemonize,
		localPort: localPort,
		stopCh:    make(chan struct{}),
	}

	s.tunnels = append(s.tunnels, tunnel)
	return tunnel
}

// Start tunnel and wait till a tcp socket is reachable
func (t *Tunnel) Start() error {
	t.stopCh = make(chan struct{})

	// ensure there is connectivity to the bastion
	bastionClient, err := t.ssh.bastionClient()
	if err != nil {
		return err
	}
	t.bastionConn = bastionClient

	if t.daemonize {
		err := t.startDaemon()
		if err != nil {
			return err
		}

		// allow for some warm up time
		time.Sleep(time.Second * 3)
		return nil
	}

	listener, err := net.Listen("tcp", net.JoinHostPort(t.BindAddress(), t.Port()))
	if err != nil {
		return err
	}
	t.listener = listener

	go t.handle()

	return nil
}

func (t *Tunnel) handle() {
	tries := 10
	for {
		remoteConn, err := t.bastionConn.Dial("tcp",
			net.JoinHostPort(t.dest, t.destPort))
		if err != nil {
			tries--
			if tries == 0 {
				return
			}

			select {
			case <-t.stopCh:
				return
			default:
			}

			time.Sleep(time.Second * 3)
			continue
		}

		t.openedConns = append(t.openedConns, remoteConn)

		conn, err := t.listener.Accept()
		if err != nil {
			select {
			case <-t.stopCh:
				return
			default:
			}

			t.log.Warnf("error accepting ssh tunnel connection: %s", err)
			continue
		}
		t.openedConns = append(t.openedConns, conn)

		go func() {
			io.Copy(remoteConn, conn)
		}()

		go func() {
			io.Copy(conn, remoteConn)
		}()
	}
}

// prevent tarmak clean up from killing used daemons
func (t *Tunnel) Stop() {
	t.cleanup()

	if t.daemon != nil {
		t.daemon.Kill()
	}
}

func (t *Tunnel) cleanup() {
	// prevent closing the same connection multiple times at once
	t.closeConnsLock.Lock()
	defer t.closeConnsLock.Unlock()

	select {
	case <-t.stopCh:
	default:
		close(t.stopCh)
	}

	for _, o := range t.openedConns {
		if o != nil {
			o.Close()
		}
	}

	if t.listener != nil {
		t.listener.Close()
	}
}

func (t *Tunnel) Port() string {
	return t.localPort
}

func (t *Tunnel) BindAddress() string {
	return "127.0.0.1"
}

func (t *Tunnel) startDaemon() error {
	binaryPath, err := osext.Executable()
	if err != nil {
		return fmt.Errorf("error finding tarmak executable: %s", err)
	}

	cmd := exec.Command(binaryPath, "tunnel", t.dest, t.destPort, t.localPort)

	outR, outW := io.Pipe()
	errR, errW := io.Pipe()
	outS := bufio.NewScanner(outR)
	errS := bufio.NewScanner(errR)

	cmd.Stdin = nil
	cmd.Stdout = outW
	cmd.Stderr = errW

	go func() {
		for outS.Scan() {
			t.log.WithField("tunnel", t.dest).Debug(outS.Text())
		}
	}()
	go func() {
		for errS.Scan() {
			t.log.WithField("tunnel", t.dest).Debug(errS.Text())
		}
	}()

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid:    true,
		Foreground: false,
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	t.daemon = cmd.Process

	return nil
}
