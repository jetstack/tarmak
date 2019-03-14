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
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	multierror "github.com/hashicorp/go-multierror"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"

	clusterv1alpha1 "github.com/jetstack/tarmak/pkg/apis/cluster/v1alpha1"
	"github.com/jetstack/tarmak/pkg/tarmak/interfaces"
)

var _ interfaces.SSH = &SSH{}

var (
	hostKeyCallbackError = errors.New("host key callback rejected")
)

const (
	// the known hosts package does not export these error strings so have to
	// define them here ourselves
	unknownKeyError  = "knownhosts: key is unknown"
	mismatchKeyError = "knownhosts: key mismatch"
)

type SSH struct {
	tarmak interfaces.Tarmak
	log    *logrus.Entry

	hosts   map[string]interfaces.Host
	tunnels []*Tunnel

	bastionClientLock sync.Mutex // ensure we set up connection to bastion one at a time to keep single connection
	bastionClientConn *ssh.Client
	bastionConfig     *ssh.ClientConfig
}

func New(tarmak interfaces.Tarmak) *SSH {
	s := &SSH{
		tarmak: tarmak,
		log:    tarmak.Log(),
	}

	return s
}

func (s *SSH) WriteConfig(c interfaces.Cluster) error {
	err := os.MkdirAll(filepath.Dir(c.SSHConfigPath()), 0700)
	if err != nil {
		return err
	}

	hosts, err := c.ListHosts()
	if err != nil {
		return err
	}

	knownHostsPath := s.tarmak.Cluster().SSHHostKeysPath()

	knownHostsFile, err := os.OpenFile(
		knownHostsPath,
		os.O_APPEND|os.O_WRONLY|os.O_CREATE,
		0600,
	)
	if err != nil {
		return err
	}
	defer knownHostsFile.Close()

	var sshConfig bytes.Buffer
	sshConfig.WriteString(fmt.Sprintf("# ssh config for tarmak cluster %s\n", c.ClusterName()))

	s.hosts = make(map[string]interfaces.Host)

	strictChecking := "yes"
	if s.tarmak.Config().IgnoreMissingPublicKeyTags() {
		strictChecking = "ask"
	}

	// loop over hosts
HOSTLOOP:
	for i := 0; i < len(hosts); i++ {
		hostKeys, err := hosts[i].SSHHostPublicKeys()
		if err != nil {
			return err
		}

		// create known hosts validator
		knownHostsValidator, err := knownhosts.New(knownHostsPath)
		if err != nil {
			return err
		}

		var errResult *multierror.Error
		// loop over host keys
		for _, hostKey := range hostKeys {
			address := &net.TCPAddr{IP: net.ParseIP(hosts[i].Hostname()), Port: 22}

			result := knownHostsValidator(
				address.String(), // empty host key
				address,          // ip address
				hostKey,
			)
			if result == nil {
				continue
			}

			e, ok := result.(*knownhosts.KeyError)

			if ok && e.Error() == mismatchKeyError {
				s.log.Warnf("removing mismatched keys for host %s, in favour of tags",
					hosts[i].Hostname())

				knownHostsFile.Close()

				b, err := ioutil.ReadFile(knownHostsPath)
				if err != nil {
					return err
				}

				knownHostsFile, err = os.Create(knownHostsPath)
				if err != nil {
					return err
				}

				for _, entry := range bytes.Split(b, []byte{'\n'}) {
					if len(entry) > 0 &&
						!bytes.Contains(entry, []byte(hosts[i].Hostname())) {

						_, err = knownHostsFile.Write(append(entry, '\n'))
						if err != nil {
							return err
						}
					}
				}

				// retry this host after removing it's keys
				i--
				continue HOSTLOOP
			}

			// if we have unknown key add to hosts file
			if ok && e.Error() == unknownKeyError {
				// add public key to known hosts file
				_, err := knownHostsFile.WriteString(
					fmt.Sprintf("%s\n", knownhosts.Line([]string{
						knownhosts.Normalize(address.String()),
					}, hostKey)),
				)

				if err != nil {
					errResult = multierror.Append(errResult,
						fmt.Errorf("failed to add public key to known hosts file: %s", err))
				}

				continue
			}

			errResult = multierror.Append(errResult,
				fmt.Errorf("ssh verification for %s failed: %v", hosts[i].Hostname(), result))
		}

		if errResult != nil {
			return errResult
		}

		_, err = sshConfig.WriteString(hosts[i].SSHConfig(strictChecking))
		if err != nil {
			return err
		}

		if len(hosts[i].Aliases()) == 0 {
			return fmt.Errorf("found host with no aliases: %s", hosts[i].Hostname())
		}

		s.hosts[hosts[i].Aliases()[0]] = hosts[i]
	}

	err = ioutil.WriteFile(c.SSHConfigPath(), sshConfig.Bytes(), 0600)
	if err != nil {
		return err
	}

	return nil
}

// Pass through a local CLI session
func (s *SSH) PassThrough(argsAdditional []string) error {
	args := append(
		[]string{
			"ssh", "-F",
			s.tarmak.Cluster().SSHConfigPath(),
		},
		argsAdditional...,
	)

	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin

	err := cmd.Start()
	if err != nil {
		return err
	}

	return cmd.Wait()
}

func (s *SSH) Execute(host string, cmd []string, stdin io.Reader, stdout, stderr io.Writer) (int, error) {
	client, err := s.client(host)
	if err != nil {
		return -1, err
	}

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

	// We don't want to default to os.Stdin here since it breaks Stdin in
	// non-obvious ways elsewhere. If os.Stdin is required here, it should be
	// passed in
	sess.Stdin = stdin

	err = sess.Start(strings.Join(cmd, " "))
	if err != nil {
		return -1, err
	}

	complete := make(chan struct{})

	go func() {
		err = sess.Wait()
		close(complete)
	}()

	select {
	case <-s.tarmak.CancellationContext().Done():
		client.Close()
		s.Cleanup()
		<-complete

	case <-complete:
	}

	if err != nil {
		if e, ok := err.(*ssh.ExitError); ok {
			return e.ExitStatus(), e
		}
		return -1, err
	}

	return 0, nil
}

func (s *SSH) client(hostName string) (*ssh.Client, error) {
	bastionClient, err := s.bastionClient()
	if err != nil {
		return nil, err
	}

	// ssh into bastion so no need to set up proxy hop
	if hostName == clusterv1alpha1.InstancePoolTypeBastion {
		return bastionClient, nil
	}

	host, err := s.host(hostName)
	if err != nil {
		return nil, err
	}

	conn, err := bastionClient.Dial("tcp", net.JoinHostPort(host.Hostname(), "22"))
	if err != nil {
		return nil, fmt.Errorf("failed to set up connection to %s from basiton: %s", host.Hostname(), err)
	}

	conf, err := s.config()
	if err != nil {
		return nil, err
	}

	ncc, chans, reqs, err := ssh.NewClientConn(conn, net.JoinHostPort(host.Hostname(), "22"), conf)
	if err != nil {
		return nil, fmt.Errorf("failed to set up ssh client: %s", err)
	}

	return ssh.NewClient(ncc, chans, reqs), nil
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

func (s *SSH) Cleanup() {
	for _, tunnel := range s.tunnels {
		tunnel.cleanup()
	}

	if s.bastionClientConn != nil {
		s.bastionClientConn.Close()
	}
}

func (s *SSH) bastionClient() (*ssh.Client, error) {
	s.bastionClientLock.Lock()
	defer s.bastionClientLock.Unlock()

	// if the current bastion client is healthy we can use this
	if s.bastionClientConn != nil {
		sess, err := s.bastionClientConn.NewSession()
		if err == nil {
			err = sess.Run("/bin/true")
			if err == nil {
				return s.bastionClientConn, nil
			}
		}

		s.log.Warnf("current connection to bastion failed: %s", err)
		s.bastionClientConn = nil
	}

	conf, err := s.config()
	if err != nil {
		return nil, err
	}

	bastion, err := s.host(clusterv1alpha1.InstancePoolTypeBastion)
	if err != nil {
		return nil, err
	}

	client, err := ssh.Dial("tcp", net.JoinHostPort(bastion.Hostname(), "22"), conf)
	if err != nil {
		return nil, fmt.Errorf("failed to set up connection to bastion: %s", err)
	}
	s.bastionClientConn = client

	s.log.Infof("new connection to bastion host successful")

	return client, nil
}

func (s *SSH) config() (*ssh.ClientConfig, error) {
	if s.bastionConfig != nil {
		return s.bastionConfig, nil
	}

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
		HostKeyCallback: s.hostKeyCallback,
	}, nil
}

func (s *SSH) hostKeyCallback(hostname string, remote net.Addr, key ssh.PublicKey) error {
	knownHostsPath := s.tarmak.Cluster().SSHHostKeysPath()

	validate := func() error {

		// create known hosts validator - refresh host file db
		knownHostsValidator, err := knownhosts.New(knownHostsPath)
		if err != nil {
			return err
		}

		return knownHostsValidator(hostname, remote, key)
	}

	err := validate()
	if err == nil {
		return nil
	}

	e, ok := err.(*knownhosts.KeyError)

	if ok && e.Error() == mismatchKeyError {
		// attempt to re-write the config if we mismatch to check tags
		if err := s.WriteConfig(s.tarmak.Cluster()); err != nil {
			return err
		}

		err = validate()
		if err == nil {
			return nil
		}

		e, ok = err.(*knownhosts.KeyError)
	}

	// the key does not exist in the known hosts file and we are ignoring
	// missing keys so append to file and retry validator
	if ok && e.Error() == unknownKeyError &&
		s.tarmak.Config().IgnoreMissingPublicKeyTags() {

		f, ferr := os.OpenFile(knownHostsPath, os.O_APPEND|os.O_WRONLY, 0600)
		if ferr != nil {
			return fmt.Errorf("failed to open file: %s, %s", ferr, err)
		}

		defer f.Close()

		entry := fmt.Sprintf("%s\n", knownhosts.Line([]string{
			knownhosts.Normalize(remote.String()),
		}, key))

		if _, ferr = f.WriteString(entry); ferr != nil {
			return fmt.Errorf("failed to append file: %s, %s", ferr, err)
		}

		return validate()
	}

	return err
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
