package config

import (
	"testing"
)

func TestStack_StackName(t *testing.T) {
	var name string
	s := &Stack{
		State: &StackState{},
	}

	name = s.StackName()
	if exp, act := StackNameState, name; exp != act {
		t.Errorf("unexpected name: exp '%s', act '%s'", exp, act)
	}

	s.Kubernetes = &StackKubernetes{}
	name = s.StackName()
	if exp, act := "", name; exp != act {
		t.Errorf("unexpected name: exp '%s', act '%s'", exp, act)
	}
}
