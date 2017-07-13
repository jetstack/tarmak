package stack

import (
	"github.com/jetstack/tarmak/pkg/tarmak/config"
	"github.com/jetstack/tarmak/pkg/tarmak/interfaces"
)

type KubernetesStack struct {
	*Stack
}

var _ interfaces.Stack = &KubernetesStack{}

func newKubernetesStack(s *Stack, conf *config.StackKubernetes) (*KubernetesStack, error) {
	s.name = config.StackNameKubernetes
	return &KubernetesStack{
		Stack: s,
	}, nil
}

func (s *KubernetesStack) Variables() map[string]interface{} {
	return map[string]interface{}{}
}

func (s *KubernetesStack) VerifyPost() error {
	return nil
}
