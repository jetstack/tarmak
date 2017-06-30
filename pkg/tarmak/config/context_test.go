package config

import (
	"strings"
	"testing"
)

func TestContext_Validate_NetworkStack(t *testing.T) {
	var c *Context

	// valid example
	c = &Context{
		Name:        "correct",
		environment: &Environment{Name: "test"},
		Stacks: []Stack{
			Stack{
				Network: &StackNetwork{
					NetworkCIDR: "1.2.3.0/24",
				},
			},
		},
	}
	if err := c.Validate(); err != nil {
		t.Errorf("unexpected error: ", err)
	}

	// two network stacks
	c = &Context{
		Name:        "twonetwork",
		environment: &Environment{Name: "test"},
		Stacks: []Stack{
			Stack{
				Network: &StackNetwork{
					NetworkCIDR: "1.2.3.0/24",
				},
			},
			Stack{
				Network: &StackNetwork{
					NetworkCIDR: "2.3.4.0/24",
				},
			},
		},
	}
	if err := c.Validate(); err == nil {
		t.Error("expect error")
	} else if contain := "multiple network stacks"; !strings.Contains(err.Error(), contain) {
		t.Errorf("expect error message '%s' to contain: '%s'", err.Error(), contain)
	}

	// no network stacks
	c = &Context{
		Name:        "nonetwork",
		environment: &Environment{Name: "test"},
		Stacks:      []Stack{},
	}
	if err := c.Validate(); err == nil {
		t.Error("expect error")
	} else if contain := "has no network stack"; !strings.Contains(err.Error(), contain) {
		t.Errorf("expect error message '%s' to contain: '%s'", err.Error(), contain)
	}
}
