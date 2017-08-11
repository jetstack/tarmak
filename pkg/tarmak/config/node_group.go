package config

import ()

type NodeGroup struct {
	Name string `yaml:"name,omitempty"`
	Role string `yaml:"role,omitempty"`

	Count int `yaml:"count,omitempty"`

	Volumes []Volume `yaml:"volumes,omitempty"`

	AWS *NodeGroupAWS `yaml:"aws,omitempty"`
}

type NodeGroupAWS struct {
	InstanceType string  `yaml:"instanceType,omitempty"`
	SpotPrice    float64 `yaml:"spotPrice,omitempty"`
}

type Volume struct {
	Name string `yaml:"name,omitempty"`
	Size int    `yaml:"size,omitempty"` // Size in GB

	AWS *VolumeAWS `yaml:"aws,omitempty"`
}

type VolumeAWS struct {
	Type string `yaml:"type,omitempty"` // gp2/st1
	// TODO: io1 (*needs more arguments) but would be good for at least etcd data dir
}
