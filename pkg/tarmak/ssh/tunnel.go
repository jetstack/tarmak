// Copyright Jetstack Ltd. See LICENSE for details.
package ssh

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os/exec"
	"strconv"
	"syscall"
	"time"

	"github.com/kardianos/osext"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"

	clusterv1alpha1 "github.com/jetstack/tarmak/pkg/apis/cluster/v1alpha1"
	"github.com/jetstack/tarmak/pkg/tarmak/interfaces"
	"github.com/jetstack/tarmak/pkg/tarmak/utils"
)

type Tunnel struct {
	log *logrus.Entry
	ssh *SSH

	dest      string
	destPort  string
	localPort string
	daemonize bool

	serverConn *ssh.Client
	listener   net.Listener

	remoteConns, localConns []net.Conn
}

var _ interfaces.Tunnel = &Tunnel{}

// This opens a local tunnel through a SSH connection
func (s *SSH) Tunnel(dest, destPort string, daemonize bool) interfaces.Tunnel {
	tunnel := &Tunnel{
		log:       s.log.WithField("destination", dest),
		ssh:       s,
		dest:      dest,
		destPort:  destPort,
		daemonize: daemonize,
		localPort: strconv.Itoa(utils.UnusedPort()),
	}

	s.tunnels = append(s.tunnels, tunnel)
	return tunnel
}

// Start tunnel and wait till a tcp socket is reachable
func (t *Tunnel) Start() error {
	// ensure there is connectivity to the bastion
	args := []string{"bastion", "/bin/true"}
	t.log.Debugf("checking SSH connection to bastion cmd=%s", args[1])
	ret, err := t.ssh.Execute(args[0], args[1:], nil, nil, nil)
	if err != nil || ret != 0 {
		return fmt.Errorf("error checking SSH connecting to bastion (%d): %s", ret, err)
	}
	t.log.Debug("connection to bastion successful")

	if t.daemonize {
		return t.startDaemon()
	}

	conf, err := t.ssh.config()
	if err != nil {
		return err
	}

	bastion, err := t.ssh.host(clusterv1alpha1.InstancePoolTypeBastion)
	if err != nil {
		return err
	}

	serverConn, err := ssh.Dial("tcp", net.JoinHostPort(bastion.Hostname(), "22"), conf)
	if err != nil {
		return err
	}
	t.serverConn = serverConn

	listener, err := net.Listen("tcp", net.JoinHostPort(t.BindAddress(), t.Port()))
	if err != nil {
		return err
	}
	t.listener = listener

	go t.handle()

	return nil
}

func (t *Tunnel) handle() {
	for {
		remoteConn, err := t.serverConn.Dial("tcp",
			net.JoinHostPort(t.dest, t.destPort))
		if err != nil {
			net.ErrWriteToConnected.
				t.log.Errorf("failed to create tunnel: %s", err)
			time.Sleep(time.Second * 2)
			continue
		}
		t.remoteConns = append(t.remoteConns, remoteConn)

		conn, err := t.listener.Accept()
		if err != nil {
			t.log.Warnf("error accepting ssh tunnel connection: %s", err)
			continue
		}
		t.localConns = append(t.localConns, conn)

		go func() {
			io.Copy(remoteConn, conn)
			remoteConn.Close()
		}()

		go func() {
			io.Copy(conn, remoteConn)
			conn.Close()
		}()
	}
}

func (t *Tunnel) Stop() {
	for _, l := range t.localConns {
		if l != nil {
			l.Close()
		}
	}
	for _, r := range t.remoteConns {
		if r != nil {
			r.Close()
		}
	}

	if t.listener != nil {
		t.listener.Close()
	}
	if t.serverConn != nil {
		t.serverConn.Close()
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

	cmd := exec.Command(binaryPath, "tunnel", t.dest, t.destPort)

	outR, outW := io.Pipe()
	errR, errW := io.Pipe()
	outS := bufio.NewScanner(outR)
	errS := bufio.NewScanner(errR)

	cmd.Stdin = nil
	cmd.Stdout = outW
	cmd.Stderr = errW

	go func() {
		for outS.Scan() {
			t.log.WithField("tunnel", t.dest).Info(outS.Text())
		}
	}()
	go func() {
		for errS.Scan() {
			t.log.WithField("tunnel", t.dest).Error(errS.Text())
		}
	}()

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid:    true,
		Foreground: false,
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	return nil
}
