package config

import ()

type Stack struct {
	State      *StackState      `yaml:"state,omitempty"`
	Network    *StackNetwork    `yaml:"network,omitempty"`
	Tools      *StackTools      `yaml:"tools,omitempty"`
	Vault      *StackVault      `yaml:"vault,omitempty"`
	Kubernetes *StackKubernetes `yaml:"kubernetes,omitempty"`
	Custom     *StackCustom     `yaml:"custom,omitempty"`
	context    *Context

	NodeGroups []NodeGroup `yaml:"nodeGroups,omitempty"`
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
}

type StackCustom struct {
	Name string `yaml:"name,omitempty"`
	Path string `yaml:"path,omitempty"`
}
