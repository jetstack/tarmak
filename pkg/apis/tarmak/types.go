package tarmak

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type Config struct {
	metav1.TypeMeta
	metav1.ObjectMeta

	CurrentContext string
	Environments   []Environment

	Contact string
	Project string
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type ConfigList struct {
	metav1.TypeMeta
	metav1.ObjectMeta

	Items []Config
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type Environment struct {
	metav1.TypeMeta
	metav1.ObjectMeta

	AWS *AWSConfig
	GCP *GCPConfig

	Contact string
	Project string

	SSHKeyPath string

	Contexts []Context
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type AWSConfig struct {
	metav1.TypeMeta
	metav1.ObjectMeta

	VaultPath         string
	AllowedAccountIDs []string
	AvailabiltyZones  []string
	Region            string
	KeyName           string
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type GCPConfig struct {
	metav1.TypeMeta
	metav1.ObjectMeta

	Zones   []string
	Region  string
	Project string
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type Context struct {
	metav1.TypeMeta
	metav1.ObjectMeta

	Stacks []Stack

	Contact string
	Project string

	BaseImage string
}

type Stack struct {
	metav1.TypeMeta

	State      *StackState
	Network    *StackNetwork
	Tools      *StackTools
	Vault      *StackVault
	Kubernetes *StackKubernetes
	Custom     *StackCustom

	NodeGroups []NodeGroup
}

type StackTools struct {
}

type StackNetwork struct {
	PeerContext string
	NetworkCIDR string
	PrivateZone string
}

type StackVault struct {
}

type StackState struct {
	BucketPrefix string
	PublicZone   string
}

type StackKubernetes struct {
}

type StackCustom struct {
	Name string
	Path string
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type NodeGroup struct {
	metav1.TypeMeta
	metav1.ObjectMeta

	Role string

	Count int

	Volumes []Volume

	AWS *NodeGroupAWS
}

type NodeGroupAWS struct {
	InstanceType string
	SpotPrice    float64
}

type Volume struct {
	Name string
	Size int

	AWS *VolumeAWS
}

type VolumeAWS struct {
	Type string
	// TODO: io1 (*needs more arguments) but would be good for at least etcd data dir
}
