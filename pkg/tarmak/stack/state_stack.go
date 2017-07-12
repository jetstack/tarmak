package stack

import (
	"github.com/jetstack/tarmak/pkg/tarmak/config"
	"github.com/jetstack/tarmak/pkg/tarmak/interfaces"
)

type StateStack struct {
	*Stack
}

var _ interfaces.Stack = &StateStack{}

func newStateStack(s *Stack, conf *config.StackState) (*StateStack, error) {
	s.name = config.StackNameState
	return &StateStack{
		Stack: s,
	}, nil
}

func (s *StateStack) Variables() map[string]interface{} {
	output := map[string]interface{}{}
	state := s.Stack.conf.State
	if state.BucketPrefix != "" {
		output["bucket_prefix"] = state.BucketPrefix
	}
	if state.PublicZone != "" {
		output["public_zone"] = state.PublicZone
	}

	return output
}
