package stack

import (
	"github.com/jetstack/tarmak/pkg/tarmak/config"
	"github.com/jetstack/tarmak/pkg/tarmak/interfaces"
)

type CustomStack struct {
	*Stack
}

var _ interfaces.Stack = &CustomStack{}

func newCustomStack(s *Stack, conf *config.StackCustom) (*CustomStack, error) {
	s.name = "custom"
	return &CustomStack{
		Stack: s,
	}, nil
}

func (s *CustomStack) Variables() map[string]interface{} {
	return map[string]interface{}{}
}

func (s *CustomStack) VerifyPost() error {
	return nil
}
