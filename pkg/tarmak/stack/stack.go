package stack

import (
	//"errors"
	"fmt"

	"github.com/Sirupsen/logrus"
	"github.com/hashicorp/go-multierror"

	tarmakv1alpha1 "github.com/jetstack/tarmak/pkg/apis/tarmak/v1alpha1"
	"github.com/jetstack/tarmak/pkg/tarmak/interfaces"
	"github.com/jetstack/tarmak/pkg/tarmak/node_group"
	"github.com/jetstack/tarmak/pkg/tarmak/role"
)

type Stack struct {
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

func New(context interfaces.Context, name string) (interfaces.Stack, error) {
	var stack interfaces.Stack
	var err error
	s := &Stack{
		context: context,
		log:     context.Log(),
	}

	// init stack
	switch name {
	case tarmakv1alpha1.StackNameState:
		stack, err = newStateStack(s)
	case tarmakv1alpha1.StackNameNetwork:
		stack, err = newNetworkStack(s)
	case tarmakv1alpha1.StackNameTools:
		stack, err = newToolsStack(s)
	case tarmakv1alpha1.StackNameVault:
		stack, err = newVaultStack(s)
	case tarmakv1alpha1.StackNameKubernetes:
		stack, err = newKubernetesStack(s)
	default:
		return nil, fmt.Errorf("unmatched state name: %s", name)
	}
	if err != nil {
		return nil, fmt.Errorf("error initialising %s stack: %s", name, err)
	}

	// init node groups
	var result error
	for _, serverPool := range context.ServerPools() {
		// see if type is handled by that stack
		if _, ok := s.roles[serverPool.Type]; !ok {
			continue
		}

		// create node groups
		nodeGroup, err := node_group.NewFromConfig(stack, &serverPool)
		if err != nil {
			result = multierror.Append(result, err)
			continue
		}
		s.nodeGroups = append(s.nodeGroups, nodeGroup)
	}

	return stack, result
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

func (s *Stack) verifyImageIDs() error {

	_, err := s.context.ImageIDs()
	if err != nil {
		return err
	}

	// TODO make sure contains my images

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

func (s *Stack) Variables() map[string]interface{} {
	vars := make(map[string]interface{})
	imageIDs, err := s.context.ImageIDs()
	if err != nil {
		s.log.Warnf("error getting image IDs: %s", err)
		return vars
	}

	for _, nodeGroup := range s.NodeGroups() {
		image := nodeGroup.Image()
		ids, ok := imageIDs[image]
		if ok {
			vars[fmt.Sprintf("%s_ami", nodeGroup.TFName())] = ids
		}
	}
	return vars

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
