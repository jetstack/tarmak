package stack

import (
	"github.com/jetstack/tarmak/pkg/tarmak/config"
	"github.com/jetstack/tarmak/pkg/tarmak/interfaces"
)

type ToolsStack struct {
	*Stack
}

var _ interfaces.Stack = &ToolsStack{}

func newToolsStack(s *Stack, conf *config.StackTools) (*ToolsStack, error) {
	s.name = config.StackNameTools
	return &ToolsStack{
		Stack: s,
	}, nil
}

func (s *ToolsStack) Variables() map[string]interface{} {
	return map[string]interface{}{}
}
