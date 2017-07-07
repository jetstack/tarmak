package config

import (
	"strings"
	"testing"
)

func testDefaultEnvironment() *Environment {
	return &Environment{
		Name: "correct",
		Contexts: []Context{
			Context{
				Name: "cluster1",
				Stacks: []Stack{
					Stack{
						State: &StackState{},
					},
					Stack{
						Network: &StackNetwork{
							NetworkCIDR: "1.2.0.0/20",
						},
					},
					Stack{
						Tools: &StackTools{},
					},
					Stack{
						Vault: &StackVault{},
					},
				},
			},
			Context{
				Name: "cluster2",
				Stacks: []Stack{
					Stack{
						Network: &StackNetwork{
							NetworkCIDR: "1.2.16.0/20",
						},
					},
				},
			},
			Context{
				Name: "cluster3",
				Stacks: []Stack{
					Stack{
						Network: &StackNetwork{
							NetworkCIDR: "1.2.32.0/20",
						},
					},
				},
			},
		},
	}
}

func TestEnvironment_Validate_NetworkOverlap(t *testing.T) {
	var e *Environment

	// valid example
	e = testDefaultEnvironment()
	if err := e.Validate(); err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	// network clash example
	e.Contexts[1].stackNetwork.Network.NetworkCIDR = "1.2.4.0/24"
	if err := e.Validate(); err == nil {
		t.Error("expect error")
	} else if contain := "network '1.2.0.0/20' overlaps with '1.2.4.0/24'"; !strings.Contains(err.Error(), contain) {
		t.Errorf("expect error message '%s' to contain: '%s'", err.Error(), contain)
	}

	// missing state
	e = testDefaultEnvironment()
	e.Contexts[0].Stacks = e.Contexts[0].Stacks[1:]
	if err := e.Validate(); err == nil {
		t.Error("expect error")
	} else if contain := "has no state stack"; !strings.Contains(err.Error(), contain) {
		t.Errorf("expect error message '%s' to contain: '%s'", err.Error(), contain)
	}

	// duplicate state
	e = testDefaultEnvironment()
	e.Contexts[0].Stacks = append(e.Contexts[0].Stacks, Stack{State: &StackState{}})
	if err := e.Validate(); err == nil {
		t.Error("expect error")
	} else if contain := "has multiple state stacks"; !strings.Contains(err.Error(), contain) {
		t.Errorf("expect error message '%s' to contain: '%s'", err.Error(), contain)
	}

	// missing tools
	e = testDefaultEnvironment()
	stacks := e.Contexts[0].Stacks[0:2]
	stacks = append(stacks, e.Contexts[0].Stacks[3:]...)
	e.Contexts[0].Stacks = stacks
	if err := e.Validate(); err == nil {
		t.Error("expect error")
	} else if contain := "has no tools stack"; !strings.Contains(err.Error(), contain) {
		t.Errorf("expect error message '%s' to contain: '%s'", err.Error(), contain)
	}

	// duplicate tools
	e = testDefaultEnvironment()
	e.Contexts[0].Stacks = append(e.Contexts[0].Stacks, Stack{Tools: &StackTools{}})
	if err := e.Validate(); err == nil {
		t.Error("expect error")
	} else if contain := "has multiple tools stacks"; !strings.Contains(err.Error(), contain) {
		t.Errorf("expect error message '%s' to contain: '%s'", err.Error(), contain)
	}

	// missing vault
	e = testDefaultEnvironment()
	stacks = e.Contexts[0].Stacks[0:3]
	stacks = append(stacks, e.Contexts[0].Stacks[4:]...)
	e.Contexts[0].Stacks = stacks
	if err := e.Validate(); err == nil {
		t.Error("expect error")
	} else if contain := "has no vault stack"; !strings.Contains(err.Error(), contain) {
		t.Errorf("expect error message '%s' to contain: '%s'", err.Error(), contain)
	}

	// duplicate vault
	e = testDefaultEnvironment()
	e.Contexts[0].Stacks = append(e.Contexts[0].Stacks, Stack{Vault: &StackVault{}})
	if err := e.Validate(); err == nil {
		t.Error("expect error")
	} else if contain := "has multiple vault stacks"; !strings.Contains(err.Error(), contain) {
		t.Errorf("expect error message '%s' to contain: '%s'", err.Error(), contain)
	}

}
