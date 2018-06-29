// Copyright Jetstack Ltd. See LICENSE for details.
package ssh

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
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
}

func New(tarmak interfaces.Tarmak) *SSH {
	s := &SSH{
		tarmak: tarmak,
		log:    tarmak.Log(),
	}

	return s
}

func (s *SSH) Validate() error {
	var result *multierror.Error

	for _, path := range []string{
		s.tarmak.Cluster().SSHConfigPath(),
		s.tarmak.Environment().SSHPrivateKeyPath(),
	} {

		f, err := os.Stat(path)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}

			result = multierror.Append(result, fmt.Errorf("failed to get '%s' file stat: %v", path, err))
			continue
		}

		if f.Mode() != os.FileMode(0600) {
			err := fmt.Errorf("'%s' does not match permissions (0600): %v", path, f.Mode())
			result = multierror.Append(result, err)
			continue
		}
	}

	return result.ErrorOrNil()
}

func (s *SSH) WriteConfig(c interfaces.Cluster) error {

	hosts, err := c.ListHosts()
	if err != nil {
		return err
	}

	var sshConfig bytes.Buffer
	sshConfig.WriteString(fmt.Sprintf("# ssh config for tarmak cluster %s\n", c.ClusterName()))

	for _, host := range hosts {
		_, err = sshConfig.WriteString(host.SSHConfig())
		if err != nil {
			return err
		}
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
	cmd.Stderr = os.Stderr
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
