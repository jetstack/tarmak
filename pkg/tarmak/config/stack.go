package config

import (
	"errors"
	"fmt"

	"github.com/hashicorp/go-multierror"
)

type Stack struct {
	State      *StackState      `yaml:"state,omitempty"`
	Network    *StackNetwork    `yaml:"network,omitempty"`
	Tools      *StackTools      `yaml:"tools,omitempty"`
	Vault      *StackVault      `yaml:"vault,omitempty"`
	Kubernetes *StackKubernetes `yaml:"kubernetes,omitempty"`
	Custom     *StackCustom     `yaml:"custom,omitempty"`
	context    *Context
}

func (s *Stack) Validate() error {
	var result error

	// ensure there is exactly one stack given
	if _, err := s.getStackName(); err != nil {
		result = multierror.Append(result, err)
	}

	return result
}

func (s *Stack) Context() *Context {
	return s.context
}

func (s *Stack) StackName() string {
	stackName, err := s.getStackName()
	if err != nil {
		return ""
	}
	return stackName
}

func (c *Stack) TerraformVars(input map[string]interface{}) map[string]interface{} {
	// state stack
	if c.StackName() == StackNameState {
		if c.State.BucketPrefix != "" {
			input["bucket_prefix"] = c.State.BucketPrefix
		}
		if c.State.PublicZone != "" {
			input["public_zone"] = c.State.PublicZone
		}
	}
	// network stack
	if c.StackName() == StackNameNetwork {
		if c.Network.NetworkCIDR != "" {
			input["network"] = c.Network.NetworkCIDR
		}
		if c.Network.PeerContext != "" {
			input["vpc_peer_stack"] = c.Network.PeerContext
		}
		if c.Network.PrivateZone != "" {
			input["private_zone"] = c.Network.PrivateZone
		}
	}
	// tools stack
	if c.StackName() == StackNameTools {
		// TODO: This is for deprecated puppet master as part of tools (get rid of that soon)
		input["puppet_deploy_key"] = "fake-ssh-key"
		input["foreman_admin_password"] = "fake-foreman-admin-password"
	}
	return input
}

func (s *Stack) getStackName() (string, error) {
	stacks := []string{}
	if s.State != nil {
		stacks = append(stacks, StackNameState)
	}
	if s.Network != nil {
		stacks = append(stacks, StackNameNetwork)
	}
	if s.Tools != nil {
		stacks = append(stacks, StackNameTools)
	}
	if s.Vault != nil {
		stacks = append(stacks, StackNameVault)
	}
	if s.Kubernetes != nil {
		stacks = append(stacks, StackNameKubernetes)
	}
	if s.Custom != nil {
		stacks = append(stacks, s.Custom.Name)
	}

	if len(stacks) < 1 {
		return "", errors.New("please specify exactly a single stack")
	}
	if len(stacks) > 1 {
		return "", fmt.Errorf("more than one stack given: %+v", stacks)
	}

	return stacks[0], nil

}

type StackTools struct {
}

type StackNetwork struct {
	PeerContext string `yaml:"peerContext,omitempty"`
	NetworkCIDR string `yaml:"networkCIDR,omitempty"`
	PrivateZone string `yaml:"privateZone,omitempty"`
}

type StackVault struct {
}

type StackState struct {
	BucketPrefix string `yaml:"bucketPrefix,omitempty"`
	PublicZone   string `yaml:"publicZone,omitempty"`
}

type StackKubernetes struct {
	EtcdCount     int     `yaml:"etcdCount,omitempty"`
	EtcdType      string  `yaml:"etcdType,omitempty"`
	EtcdSpotPrice float32 `yaml:"etcdSpotPrice,omitempty"`

	WorkerCount     int     `yaml:"workerCount,omitempty"`
	WorkerType      string  `yaml:"workerType,omitempty"`
	WorkerSpotPrice float32 `yaml:"workerSpotPrice,omitempty"`

	MasterCount     int     `yaml:"masterCount,omitempty"`
	MasterType      string  `yaml:"masterType,omitempty"`
	MasterSpotPrice float32 `yaml:"masterSpotPrice,omitempty"`
}

type StackCustom struct {
	Name string `yaml:"name,omitempty"`
	Path string `yaml:"path,omitempty"`
}
