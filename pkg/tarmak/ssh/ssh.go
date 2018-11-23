// Copyright Jetstack Ltd. See LICENSE for details.
package ssh

import (
	"bytes"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/terminal"

	clusterv1alpha1 "github.com/jetstack/tarmak/pkg/apis/cluster/v1alpha1"
	"github.com/jetstack/tarmak/pkg/tarmak/interfaces"
	"github.com/jetstack/tarmak/pkg/tarmak/utils"
)

var _ interfaces.SSH = &SSH{}

type SSH struct {
	tarmak interfaces.Tarmak
	log    *logrus.Entry

	hosts   map[string]interfaces.Host
	tunnels []interfaces.Tunnel
}

func New(tarmak interfaces.Tarmak) *SSH {
	s := &SSH{
		tarmak: tarmak,
		log:    tarmak.Log(),
	}

	return s
}

func (s *SSH) WriteConfig(c interfaces.Cluster) error {
	hosts, err := c.ListHosts()
	if err != nil {
		return err
	}

	var sshConfig bytes.Buffer
	sshConfig.WriteString(fmt.Sprintf("# ssh config for tarmak cluster %s\n", c.ClusterName()))

	s.hosts = make(map[string]interfaces.Host)
	for _, host := range hosts {
		_, err = sshConfig.WriteString(host.SSHConfig())
		if err != nil {
			return err
		}

		if len(host.Aliases()) == 0 {
			return fmt.Errorf("found host with no aliases: %s", host.Hostname())
		}

		s.hosts[host.Aliases()[0]] = host
	}

	err = utils.EnsureDirectory(filepath.Dir(c.SSHConfigPath()), 0700)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(c.SSHConfigPath(), sshConfig.Bytes(), 0600)
	if err != nil {
		return err
	}

	return nil
}

// Pass through a local CLI session
func (s *SSH) PassThrough(hostName string, argsAdditional []string) error {
	if len(argsAdditional) > 0 {
		_, err := s.Execute(hostName, argsAdditional, nil, nil, nil)
		return err
	}

	client, err := s.client(hostName)
	if err != nil {
		return err
	}
	defer client.Close()

	sess, err := client.NewSession()
	if err != nil {
		return err
	}
	defer sess.Close()

	sess.Stderr = os.Stderr
	sess.Stdout = os.Stdout
	sess.Stdin = os.Stdin

	modes := ssh.TerminalModes{
		ssh.ECHO:          1,
		ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
	}

	fileDescriptor := int(os.Stdin.Fd())
	if terminal.IsTerminal(fileDescriptor) {
		originalState, err := terminal.MakeRaw(fileDescriptor)
		if err != nil {
			return err
		}
		defer terminal.Restore(fileDescriptor, originalState)

		termWidth, termHeight, err := terminal.GetSize(fileDescriptor)
		if err != nil {
			return err
		}

		err = sess.RequestPty("xterm-256color", termHeight, termWidth, modes)
		if err != nil {
			return err
		}
	}

	if err := sess.Shell(); err != nil {
		return err
	}

	if err := sess.Wait(); err != nil {
		return err
	}

	return nil
}

func (s *SSH) Execute(host string, cmd []string, stdin io.Reader, stdout, stderr io.Writer) (int, error) {
	client, err := s.client(host)
	if err != nil {
		return -1, err
	}
	defer client.Close()

	sess, err := client.NewSession()
	if err != nil {
		return -1, err
	}
	defer sess.Close()

	if stderr == nil {
		sess.Stderr = os.Stderr
	} else {
		sess.Stderr = stderr
	}

	if stdout == nil {
		sess.Stdout = os.Stdout
	} else {
		sess.Stdout = stdout
	}

	if stdin == nil {
		sess.Stdin = os.Stdin
	} else {
		sess.Stdin = stdin
	}

	err = sess.Start(strings.Join(cmd, " "))
	if err != nil {
		return -1, err
	}

	if err := sess.Wait(); err != nil {
		if e, ok := err.(*ssh.ExitError); ok {
			return e.ExitStatus(), e
		}
		return -1, err
	}

	return 0, nil
}

func (s *SSH) Validate() error {
	// no environment in tarmak so we have no SSH to validate
	if s.tarmak.Environment() == nil {
		return nil
	}

	keyPath := s.tarmak.Environment().SSHPrivateKeyPath()
	f, err := os.Stat(keyPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}

		return fmt.Errorf("failed to read ssh file status: %v", err)
	}

	if f.IsDir() {
		return fmt.Errorf("expected ssh file location '%s' is directory", keyPath)
	}

	if f.Mode() != os.FileMode(0600) && f.Mode() != os.FileMode(0400) {
		s.log.Warnf("ssh file '%s' holds incorrect permissions (%v), setting to 0600", keyPath, f.Mode())
		if err := os.Chmod(keyPath, os.FileMode(0600)); err != nil {
			return fmt.Errorf("failed to set ssh private key file permissions: %v", err)
		}
	}

	bytes, err := ioutil.ReadFile(keyPath)
	if err != nil {
		return fmt.Errorf("unable to read ssh private key: %s", err)
	}

	block, _ := pem.Decode(bytes)
	if block == nil {
		return errors.New("failed to parse PEM block containing the ssh private key")
	}

	return nil
}

func (s *SSH) client(hostName string) (*ssh.Client, error) {
	conf, err := s.config()
	if err != nil {
		return nil, err
	}

	bastion, err := s.host(clusterv1alpha1.InstancePoolTypeBastion)
	if err != nil {
		return nil, err
	}

	proxyClient, err := ssh.Dial("tcp", net.JoinHostPort(bastion.Hostname(), "22"), conf)
	if err != nil {
		return nil, fmt.Errorf("failed to set up connection to bastion: %s", err)
	}

	// ssh into bastion so no need to set up proxy hop
	if hostName == clusterv1alpha1.InstancePoolTypeBastion {
		return proxyClient, nil
	}

	host, err := s.host(hostName)
	if err != nil {
		return nil, err
	}

	conn, err := proxyClient.Dial("tcp", net.JoinHostPort(host.Hostname(), "22"))
	if err != nil {
		return nil, fmt.Errorf("failed to set up connection to %s from basiton: %s", host.Hostname(), err)
	}

	ncc, chans, reqs, err := ssh.NewClientConn(conn, net.JoinHostPort(host.Hostname(), "22"), conf)
	if err != nil {
		return nil, fmt.Errorf("failed to set up ssh client: %s", err)
	}

	return ssh.NewClient(ncc, chans, reqs), nil
}

func (s *SSH) config() (*ssh.ClientConfig, error) {
	bastion, err := s.host(clusterv1alpha1.InstancePoolTypeBastion)
	if err != nil {
		return nil, err
	}

	b, err := ioutil.ReadFile(s.tarmak.Environment().SSHPrivateKeyPath())
	if err != nil {
		return nil, fmt.Errorf("failed to read ssh private key: %s", err)
	}

	signer, err := ssh.ParsePrivateKey(b)
	if err != nil {
		return nil, fmt.Errorf("failed to parse ssh private key: %s", err)
	}

	return &ssh.ClientConfig{
		Timeout:         time.Minute * 10,
		User:            bastion.User(),
		Auth:            []ssh.AuthMethod{ssh.PublicKeys(signer)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}, nil
}

func (s *SSH) host(name string) (interfaces.Host, error) {
	host, ok := s.hosts[name]
	if ok {
		return host, nil
	}

	// we have already have all hosts, we can't find it
	if len(s.hosts) > 0 {
		return nil, fmt.Errorf("failed to resolve host: %s", name)
	}

	err := s.WriteConfig(s.tarmak.Cluster())
	if err != nil {
		return nil, err
	}

	_, bok := s.hosts[clusterv1alpha1.InstancePoolTypeBastion]
	err = fmt.Errorf("failed to resolve target hosts for ssh: found %s=%v",
		clusterv1alpha1.InstancePoolTypeBastion,
		bok)
	if !bok && name == clusterv1alpha1.InstancePoolTypeBastion {
		return nil, err
	}

	host, hok := s.hosts[name]
	if !hok {
		return nil, fmt.Errorf("%s %s=%v", err, name, hok)
	}

	return host, nil
}

func (s *SSH) Cleanup() {
	for _, tunnel := range s.tunnels {
		tunnel.Stop()
	}
}
