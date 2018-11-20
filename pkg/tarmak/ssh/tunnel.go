// Copyright Jetstack Ltd. See LICENSE for details.
package ssh

import (
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"time"

	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"

	"github.com/jetstack/tarmak/pkg/tarmak/interfaces"
	"github.com/jetstack/tarmak/pkg/tarmak/utils"
)

type Tunnel struct {
	localPort int
	log       *logrus.Entry
	stdin     io.WriteCloser

	retryCount int
	retryWait  time.Duration

	destinaton     string
	destinatonPort int

	forwardSpec string
	sshCommand  []string
	ssh         *SSH

	serverConn *ssh.Client
	listener   net.Listener

	remoteConns, localConns []net.Conn
}

var _ interfaces.Tunnel = &Tunnel{}

// This opens a local tunnel through a SSH connection
func (s *SSH) Tunnel(hostname string, destination string, destinationPort int) interfaces.Tunnel {
	return &Tunnel{
		localPort:      utils.UnusedPort(),
		log:            s.log.WithField("destination", destination),
		ssh:            s,
		destinaton:     destination,
		destinatonPort: destinationPort,
	}
}

// Start tunnel and wait till a tcp socket is reachable
func (t *Tunnel) Start() error {
	var err error

	// ensure there is connectivity to the bastion
	args := []string{"bastion", "/bin/true"}
	t.log.Debugf("checking SSH connection to bastion cmd=%s", args[1])
	ret, err := t.ssh.Execute(args[0], args[1:], nil, nil, nil)
	if err != nil || ret != 0 {
		return fmt.Errorf("error checking SSH connecting to bastion (%d): %s", ret, err)
	}

	b, err := ioutil.ReadFile(t.ssh.tarmak.Environment().SSHPrivateKeyPath())
	if err != nil {
		return fmt.Errorf("failed to read ssh private key: %s", err)
	}

	signer, err := ssh.ParsePrivateKey(b)
	if err != nil {
		return fmt.Errorf("failed to parse ssh private key: %s", err)
	}

	confProxy := &ssh.ClientConfig{
		Timeout:         time.Minute * 10,
		User:            t.ssh.bastion.User(),
		Auth:            []ssh.AuthMethod{ssh.PublicKeys(signer)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	serverConn, err := ssh.Dial("tcp", net.JoinHostPort(t.ssh.bastion.Hostname(), "22"), confProxy)
	if err != nil {
		return err
	}
	t.serverConn = serverConn

	listener, err := net.Listen("tcp", net.JoinHostPort(t.BindAddress(), fmt.Sprintf("%d", t.Port())))
	if err != nil {
		return err
	}
	t.listener = listener

	go t.pass()

	return nil
}

func (t *Tunnel) pass() {
	for {
		remoteConn, err := t.serverConn.Dial("tcp", net.JoinHostPort(t.destinaton, fmt.Sprintf("%d", t.destinatonPort)))
		if err != nil {
			fmt.Errorf("%s\n", err)
			return
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
		l.Close()
	}
	for _, r := range t.remoteConns {
		r.Close()
	}

	t.listener.Close()
	t.serverConn.Close()
}

func (t *Tunnel) Port() int {
	return t.localPort
}

func (t *Tunnel) BindAddress() string {
	return "127.0.0.1"
}
