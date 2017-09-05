package stack

import (
	"fmt"

	clusterv1alpha1 "github.com/jetstack/tarmak/pkg/apis/cluster/v1alpha1"
	"github.com/jetstack/tarmak/pkg/tarmak/interfaces"
	"github.com/jetstack/tarmak/pkg/tarmak/role"
)

type ToolsStack struct {
	*Stack
}

var _ interfaces.Stack = &ToolsStack{}

func newToolsStack(s *Stack) (*ToolsStack, error) {
	t := &ToolsStack{
		Stack: s,
	}

	jenkinsRole := &role.Role{
		Stateful: true,
		AWS:      &role.RoleAWS{},
	}
	jenkinsRole.WithName("jenkins")

	bastionRole := &role.Role{
		Stateful: true,
		AWS:      &role.RoleAWS{},
	}
	bastionRole.WithName("bastion")

	s.roles = map[string]*role.Role{
		clusterv1alpha1.ServerPoolTypeJenkins: jenkinsRole,
		clusterv1alpha1.ServerPoolTypeBastion: bastionRole,
	}

	s.name = StackNameTools
	s.verifyPostDeploy = append(s.verifyPostDeploy, t.verifyBastionAvailable)
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
