package stack

import (
	"fmt"

	"github.com/jetstack/tarmak/pkg/tarmak/config"
	"github.com/jetstack/tarmak/pkg/tarmak/interfaces"
)

type ToolsStack struct {
	*Stack
}

var _ interfaces.Stack = &ToolsStack{}

func newToolsStack(s *Stack, conf *config.StackTools) (*ToolsStack, error) {
	t := &ToolsStack{
		Stack: s,
	}

	s.name = config.StackNameTools
	s.verifyPost = append(s.verifyPost, t.verifyBastionAvailable)
	return t, nil
}

func (s *ToolsStack) Variables() map[string]interface{} {
	return map[string]interface{}{}
}

func (s *ToolsStack) verifyBastionAvailable() error {

	ssh := s.Context().Environment().Tarmak().SSH()

	if err := ssh.WriteConfig(); err != nil {
		return err
	}

	retCode, err := ssh.Execute(
		"bastion",
		"/bin/true",
		[]string{},
	)

	msg := "error while connectioning to bastion host"
	if err != nil {
		return fmt.Errorf("%s: %s", msg, err)
	}
	if retCode != 0 {
		return fmt.Errorf("%s unexpected return code: %d", msg, retCode)
	}

	return nil

}
