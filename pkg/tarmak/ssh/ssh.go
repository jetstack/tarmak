// Copyright Jetstack Ltd. See LICENSE for details.
package ssh

import (
	"bytes"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/hashicorp/go-multierror"
	"github.com/sirupsen/logrus"

	"github.com/jetstack/tarmak/pkg/tarmak/interfaces"
	"github.com/jetstack/tarmak/pkg/tarmak/utils"
)

var _ interfaces.SSH = &SSH{}

type SSH struct {
	tarmak interfaces.Tarmak
	log    *logrus.Entry

	controlPaths []string
}

func New(tarmak interfaces.Tarmak) *SSH {
	s := &SSH{
		tarmak: tarmak,
		log:    tarmak.Log(),
	}

	return s
}

func (s *SSH) WriteConfig(c interfaces.Cluster, interactive bool) error {
	err := utils.EnsureDirectory(filepath.Dir(c.SSHConfigPath()), 0700)
	if err != nil {
		return err
	}

	hosts, err := c.ListHosts()
	if err != nil {
		return err
	}

	localKnownHosts, err := s.parseKnownHosts()
	if err != nil {
		return err
	}

	knownHosts, err := os.OpenFile(s.tarmak.Cluster().SSHHostKeysPath(),
		os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return err
	}
	defer knownHosts.Close()

	var sshConfig bytes.Buffer
	sshConfig.WriteString(fmt.Sprintf("# ssh config for tarmak cluster %s\n", c.ClusterName()))

	for _, host := range hosts {
		strictChecking := "yes"

		if _, ok := localKnownHosts[host.Hostname()]; !ok {
			// local host key is missing, so append
			entry, err := host.SSHKnownHostConfig()
			if err != nil {
				return err
			}

			if entry == "" && s.tarmak.Config().IgnoreMissingPublicKeyTags() {
				// We need to change strict 'yes' to 'no' or 'ask' since entry doesn't
				// exist and we have 'ignore missing instances tags' set to true.  Set
				// to no for programmatic uses of ssh which will be changed when the in
				// package solution is used.

				if interactive {
					strictChecking = "ask"
				} else {
					strictChecking = "no"
				}
			}

			if _, err := knownHosts.WriteString(entry); err != nil {
				return err
			}
		}

		_, err = sshConfig.WriteString(host.SSHConfig(strictChecking))
		if err != nil {
			return err
		}

		s.controlPaths = append(s.controlPaths, host.SSHControlPath())
	}

	err = ioutil.WriteFile(c.SSHConfigPath(), sshConfig.Bytes(), 0600)
	if err != nil {
		return err
	}

	return nil
}

func (s *SSH) parseKnownHosts() (map[string]string, error) {
	b, err := ioutil.ReadFile(s.tarmak.Cluster().SSHHostKeysPath())
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	entries := make(map[string]string)
	for _, entry := range strings.Split(string(b), "\n") {
		line := strings.SplitN(entry, " ", 2)

		if len(line) == 2 {
			entries[line[0]] = line[1]
		}
	}

	return entries, nil
}

func (s *SSH) args() []string {
	return []string{
		"ssh",
		"-F",
		s.tarmak.Cluster().SSHConfigPath(),
	}
}

// Pass through a local CLI session
func (s *SSH) PassThrough(argsAdditional []string) {
	args := append(s.args(), argsAdditional...)

	cmd := exec.Command(args[0], args[1:len(args)]...)
	cmd.Stderr = os.Stdout
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin

	err := cmd.Start()
	if err != nil {
		s.log.Fatal(err)
	}

	err = cmd.Wait()
	if err != nil {
		s.log.Fatal(err)
	}
}

func (s *SSH) Execute(host string, command string, argsAdditional []string) (returnCode int, err error) {
	args := append(s.args(), host, "--", command)
	args = append(args, argsAdditional...)

	cmd := exec.Command(args[0], args[1:len(args)]...)

	err = cmd.Start()
	if err != nil {
		return -1, err
	}

	err = cmd.Wait()
	if err != nil {
		perr, ok := err.(*exec.ExitError)
		if ok {
			if status, ok := perr.Sys().(syscall.WaitStatus); ok {
				return status.ExitStatus(), nil
			}
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

func (s *SSH) Cleanup() error {
	var result *multierror.Error

	for _, c := range utils.RemoveDuplicateStrings(s.controlPaths) {
		if err := os.RemoveAll(c); err != nil {
			result = multierror.Append(result, err)
		}
	}

	return result.ErrorOrNil()
}
