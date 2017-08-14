package stack

import (
	"errors"
	"fmt"

	"github.com/Sirupsen/logrus"
	"github.com/hashicorp/go-multierror"

	"github.com/jetstack/tarmak/pkg/tarmak/config"
	"github.com/jetstack/tarmak/pkg/tarmak/interfaces"
	"github.com/jetstack/tarmak/pkg/tarmak/node_group"
	"github.com/jetstack/tarmak/pkg/tarmak/role"
)

type Stack struct {
	conf *config.Stack

	name    string
	context interfaces.Context
	log     *logrus.Entry

	verifyPreDeploy   []func() error
	verifyPreDestroy  []func() error
	verifyPostDeploy  []func() error
	verifyPostDestroy []func() error

	output map[string]interface{}

	roles map[string]*role.Role

	nodeGroups []interfaces.NodeGroup
}

func NewFromConfig(context interfaces.Context, conf *config.Stack) (interfaces.Stack, error) {
	stacks := []interfaces.Stack{}

	s := &Stack{
		conf:    conf,
		context: context,
		log:     context.Log(),
	}

	if conf.State != nil {
		stack, err := newStateStack(s, conf.State)
		if err != nil {
			return nil, fmt.Errorf("error initialising state stack: %s", err)
		}
		stacks = append(stacks, stack)
	}
	if conf.Network != nil {
		stack, err := newNetworkStack(s, conf.Network)
		if err != nil {
			return nil, fmt.Errorf("error initialising network stack: %s", err)
		}
		stacks = append(stacks, stack)
	}
	if conf.Tools != nil {
		stack, err := newToolsStack(s, conf.Tools)
		if err != nil {
			return nil, fmt.Errorf("error initialising tools stack: %s", err)
		}
		stacks = append(stacks, stack)
	}
	if conf.Vault != nil {
		stack, err := newVaultStack(s, conf.Vault)
		if err != nil {
			return nil, fmt.Errorf("error initialising vault stack: %s", err)
		}
		stacks = append(stacks, stack)
	}
	if conf.Kubernetes != nil {
		stack, err := newKubernetesStack(s, conf.Kubernetes)
		if err != nil {
			return nil, fmt.Errorf("error initialising kubernetes stack: %s", err)
		}
		stacks = append(stacks, stack)
	}
	if conf.Custom != nil {
		stack, err := newCustomStack(s, conf.Custom)
		if err != nil {
			return nil, fmt.Errorf("error initialising custom stack: %s", err)
		}
		stacks = append(stacks, stack)
	}

	if len(stacks) < 1 {
		return nil, errors.New("please specify exactly a single stack")
	}
	if len(stacks) > 1 {
		return nil, fmt.Errorf("more than one stack given: %+v", stacks)
	}

	// initialiase node groups
	var result error
	for pos, _ := range conf.NodeGroups {
		nodeGroup, err := node_group.NewFromConfig(stacks[0], &conf.NodeGroups[pos])
		if err != nil {
			result = multierror.Append(result, err)
			continue
		}
		s.nodeGroups = append(s.nodeGroups, nodeGroup)
	}

	return stacks[0], result

}

func (s *Stack) SetOutput(in map[string]interface{}) {
	s.output = in
}

func (s *Stack) Output() map[string]interface{} {
	return s.output
}

func (s *Stack) Context() interfaces.Context {
	return s.context
}

func (s *Stack) RemoteState() string {
	return s.Context().RemoteState(s.Name())
}

func (s *Stack) Name() string {
	return s.name
}

func (s *Stack) Validate() error {
	return nil
}

func (s *Stack) VerifyPreDeploy() error {
	var result error
	for _, f := range s.verifyPreDeploy {
		err := f()
		if err != nil {
			result = multierror.Append(result, err)
		}
	}
	return result
}

func (s *Stack) VerifyPreDestroy() error {
	var result error
	for _, f := range s.verifyPreDestroy {
		err := f()
		if err != nil {
			result = multierror.Append(result, err)
		}
	}
	return result
}

func (s *Stack) VerifyPostDeploy() error {
	var result error
	for _, f := range s.verifyPostDeploy {
		err := f()
		if err != nil {
			result = multierror.Append(result, err)
		}
	}
	return result
}

func (s *Stack) VerifyPostDestroy() error {
	var result error
	for _, f := range s.verifyPostDestroy {
		err := f()
		if err != nil {
			result = multierror.Append(result, err)
		}
	}
	return result
}

func (s *Stack) Log() *logrus.Entry {
	return s.log
}

func (s *Stack) Role(roleName string) *role.Role {
	if s.roles != nil {
		if role, ok := s.roles[roleName]; ok {
			return role
		}
	}
	return nil
}

func (s *Stack) Roles() (roles []*role.Role) {
	roleMap := map[string]bool{}
	for _, nodeGroup := range s.NodeGroups() {
		r := nodeGroup.Role()
		if _, ok := roleMap[r.Name()]; !ok {
			roles = append(roles, r)
			roleMap[r.Name()] = true
		}
	}
	return roles
}

func (s *Stack) NodeGroups() (nodeGroups []interfaces.NodeGroup) {
	return s.nodeGroups
}
