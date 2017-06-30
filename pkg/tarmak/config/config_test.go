package config

import (
	"strings"
	"testing"

	"gopkg.in/yaml.v2"
)

func TestStack_StackName(t *testing.T) {
	s := &Stack{
		State: &StackState{},
	}

	name, err := s.StackName()
	if err != nil {
		t.Error("unexpected error: ", err)
	}

	if exp, act := StackNameState, name; exp != act {
		t.Errorf("unexpected name: exp '%s', act '%s'", exp, act)
	}

	s.Kubernetes = &StackKubernetes{}
	_, err = s.StackName()
	if err == nil {
		t.Error("expected error when two stacks supplied")
	}

}

func TestDefaultConfigOmitEmpty(t *testing.T) {
	c := DefaultConfigHub()

	y, err := yaml.Marshal(c)
	if err != nil {
		t.Error("unexpected error: ", err)
	}

	if strings.Contains(string(y), "null") || strings.Contains(string(y), "\"\"") {
		t.Error("yaml contains empty values, probably forgot omitempty")
	}

	c = DefaultConfigSingle()

	y, err = yaml.Marshal(c)
	if err != nil {
		t.Error("unexpected error: ", err)
	}

	if strings.Contains(string(y), "null") || strings.Contains(string(y), "\"\"") {
		t.Error("yaml contains empty values, probably forgot omitempty")
	}

}

func TestDefaultConfig_Validate(t *testing.T) {
	c := DefaultConfigHub()

	err := c.Validate()
	if err != nil {
		t.Error("unexpected error: ", err)
	}

	c = DefaultConfigSingle()
	err = c.Validate()
	if err != nil {
		t.Error("unexpected error: ", err)
	}

}

func TestConfig_ValidateNetworkOverlap(t *testing.T) {

	// this should validate as the overlapping ared in different zones
	c := &Config{
		Contexts: []Context{
			Context{
				Name:        "cluster1",
				Environment: "env1",
				Stacks: []Stack{
					Stack{
						Network: &StackNetwork{
							NetworkCIDR: "10.98.0.0/20",
						},
					},
				},
			},
			Context{
				Name:        "cluster2",
				Environment: "env1",
				Stacks: []Stack{
					Stack{
						Network: &StackNetwork{
							NetworkCIDR: "10.98.64.0/20",
						},
					},
				},
			},
			Context{
				Name:        "cluster3",
				Environment: "env1",
				Stacks: []Stack{
					Stack{
						Network: &StackNetwork{
							NetworkCIDR: "10.98.128.0/20",
						},
					},
				},
			},
			Context{
				Name:        "cluster1",
				Environment: "env2",
				Stacks: []Stack{
					Stack{
						Network: &StackNetwork{
							NetworkCIDR: "10.98.0.0/20",
						},
					},
				},
			},
		},
	}

	err := c.validateNetworkOverlap()
	if err != nil {
		t.Error("unexpected error: ", err)
	}

	c.Contexts[1].Stacks[0].Network.NetworkCIDR = "10.98.1.0/24"
	err = c.validateNetworkOverlap()
	if err == nil {
		t.Error("expected error")
	} else if !strings.Contains(err.Error(), "network overlap in environment") {
		t.Error("unexpected error message: ", err)
	}
}
