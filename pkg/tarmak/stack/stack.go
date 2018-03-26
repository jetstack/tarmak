// Copyright Jetstack Ltd. See LICENSE for details.
package stack

import (
	"fmt"

	"github.com/hashicorp/go-multierror"
	"github.com/sirupsen/logrus"

	tarmakv1alpha1 "github.com/jetstack/tarmak/pkg/apis/tarmak/v1alpha1"
	"github.com/jetstack/tarmak/pkg/tarmak/interfaces"
	"github.com/jetstack/tarmak/pkg/tarmak/role"
)

type Stack struct {
	name    string
	cluster interfaces.Cluster
	log     *logrus.Entry

	verifyPreDeploy   []func() error
	verifyPreDestroy  []func() error
	verifyPostDeploy  []func() error
	verifyPostDestroy []func() error

	output map[string]interface{}

	roles map[string]bool

	instancePools []interfaces.InstancePool
}

func New(cluster interfaces.Cluster, name string) (interfaces.Stack, error) {
	var stack interfaces.Stack
	var err error
	s := &Stack{
		cluster: cluster,
		log:     cluster.Log().WithField("stack", name),
	}

	// init stack
	switch name {
	case tarmakv1alpha1.StackNameState:
		stack, err = newStateStack(s)
	case tarmakv1alpha1.StackNameNetwork:
		stack, err = newNetworkStack(s)
	case tarmakv1alpha1.StackNameExistingNetwork:
		stack, err = newExistingNetworkStack(s)
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

	return stack, nil
}

func (s *Stack) SetOutput(in map[string]interface{}) {
	s.output = in
}

func (s *Stack) Output() map[string]interface{} {
	return s.output
}

func (s *Stack) Cluster() interfaces.Cluster {
	return s.cluster
}

func (s *Stack) RemoteState() string {
	return s.Cluster().RemoteState(s.Name())
}

func (s *Stack) Name() string {
	return s.name
}

func (s *Stack) Validate() error {
	return nil
}

func (s *Stack) verifyImageIDs() error {

	_, err := s.cluster.ImageIDs()
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
	imageIDs, err := s.cluster.ImageIDs()
	if err != nil {
		s.log.Warnf("error getting image IDs: %s", err)
		return vars
	}

	for _, instancePool := range s.InstancePools() {
		image := instancePool.Image()
		ids, ok := imageIDs[image]
		if ok {
			vars[fmt.Sprintf("%s_ami", instancePool.TFName())] = ids
		}
		vars[fmt.Sprintf("%s_instance_count", instancePool.TFName())] = instancePool.Config().MinCount
	}
	return vars

}

func (s *Stack) Roles() (roles []*role.Role) {
	roleMap := map[string]bool{}
	for _, instancePool := range s.InstancePools() {
		r := instancePool.Role()
		if _, ok := roleMap[r.Name()]; !ok {
			roles = append(roles, r)
			roleMap[r.Name()] = true
		}
	}
	return roles
}

func (s *Stack) InstancePools() (instancePools []interfaces.InstancePool) {
	for _, ng := range s.cluster.InstancePools() {
		if s.roles != nil {
			if active, ok := s.roles[ng.Role().Name()]; !ok || !active {
				continue
			}
		}
		instancePools = append(instancePools, ng)
	}
	return instancePools
}
