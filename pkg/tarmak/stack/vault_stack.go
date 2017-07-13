package stack

import (
	"github.com/jetstack/tarmak/pkg/tarmak/config"
	"github.com/jetstack/tarmak/pkg/tarmak/interfaces"
)

type VaultStack struct {
	*Stack
}

var _ interfaces.Stack = &VaultStack{}

func newVaultStack(s *Stack, conf *config.StackVault) (*VaultStack, error) {
	s.name = config.StackNameVault
	return &VaultStack{
		Stack: s,
	}, nil
}

func (s *VaultStack) Variables() map[string]interface{} {
	return map[string]interface{}{}
}

func (s *VaultStack) VerifyPost() error {
	return nil
}
