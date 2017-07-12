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
