package stack

import (
	"fmt"
	"time"

	"golang.org/x/crypto/ssh"

	"github.com/jetstack/tarmak/pkg/tarmak/config"
	"github.com/jetstack/tarmak/pkg/tarmak/interfaces"
)

type ToolsStack struct {
	*Stack
}

var _ interfaces.Stack = &ToolsStack{}

func newToolsStack(s *Stack, conf *config.StackTools) (*ToolsStack, error) {
	s.name = config.StackNameTools
	return &ToolsStack{
		Stack: s,
	}, nil
}

func (s *ToolsStack) Variables() map[string]interface{} {
	return map[string]interface{}{}
}

func (s *ToolsStack) VerifyPost() error {
	return s.verifyBastionAvailable()
}

func (s *ToolsStack) verifyBastionAvailable() error {

	hosts, err := s.Context().Environment().Provider().ListHosts()
	if err != nil {
		return err
	}

	var bastionHost interfaces.Host
	for _, host := range hosts {
		for _, role := range host.Roles() {
			if role == "bastion" {
				bastionHost = host
			}
		}
	}

	signer, err := ssh.NewSignerFromKey(s.Context().Environment().SSHPrivateKey())
	if err != nil {
		return err
	}

	sshConfig := &ssh.ClientConfig{
		User: bastionHost.User(),
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		Timeout: 10 * time.Second,
		// TODO: Do this properly
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	_, err = ssh.Dial("tcp", fmt.Sprintf("%s:%d", bastionHost.Hostname(), 22), sshConfig)
	if err != nil {
		return fmt.Errorf("failed to connect to bastion: %s", err)
	}

	return nil

	/*
		key := s.Context().Environment().SSHPrivateKey()
		return nil
	*/
}
