package stack

import (
	"errors"
	"fmt"

	"github.com/Sirupsen/logrus"

	"github.com/jetstack/tarmak/pkg/tarmak/config"
	"github.com/jetstack/tarmak/pkg/tarmak/interfaces"
)

type Stack struct {
	conf *config.Stack

	name    string
	context interfaces.Context
	log     *logrus.Entry

	output map[string]interface{}
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

	return stacks[0], nil

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

func (s *Stack) Log() *logrus.Entry {
	return s.log
}
