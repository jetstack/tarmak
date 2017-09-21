package stack

import (
	"fmt"

	clusterv1alpha1 "github.com/jetstack/tarmak/pkg/apis/cluster/v1alpha1"
	tarmakv1alpha1 "github.com/jetstack/tarmak/pkg/apis/tarmak/v1alpha1"
	"github.com/jetstack/tarmak/pkg/tarmak/interfaces"
)

type ToolsStack struct {
	*Stack
}

var _ interfaces.Stack = &ToolsStack{}

func newToolsStack(s *Stack) (*ToolsStack, error) {
	t := &ToolsStack{
		Stack: s,
	}

	s.roles = make(map[string]bool)
	s.roles[clusterv1alpha1.InstancePoolTypeJenkins] = true
	s.roles[clusterv1alpha1.InstancePoolTypeBastion] = true

	s.name = tarmakv1alpha1.StackNameTools
	s.verifyPreDeploy = append(s.verifyPostDeploy, s.verifyImageIDs)
	s.verifyPostDeploy = append(s.verifyPostDeploy, t.verifyBastionAvailable)
	return t, nil
}

func (s *ToolsStack) Variables() map[string]interface{} {
	return s.Stack.Variables()
}

func (s *ToolsStack) verifyBastionAvailable() error {

	ssh := s.Cluster().Environment().Tarmak().SSH()

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
