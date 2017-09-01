package v1alpha1

import (
	clusterv1alpha1 "github.com/jetstack/tarmak/pkg/apis/cluster/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +resource:path=configs

type Config struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	CurrentContext string        `json:"currentContext,omitempty"` // <environmentName>-<contextName>
	Environments   []Environment `json:"environments,omitempty"`

	Contact string `json:"contact,omitempty"`
	Project string `json:"project,omitempty"`

	Clusters []*clusterv1alpha1.Cluster `json:"project,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type ConfigList struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Items []Config `json:"items"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type Environment struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	AWS *AWSConfig `json:"aws,omitempty"`
	GCP *GCPConfig `json:"gcp,omitempty"`

	Contact string `json:"contact,omitempty"`
	Project string `json:"project,omitempty"`

	SSHKeyPath string `json:"sshKeyPath,omitempty"`

	Contexts []Context `json:"contexts,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type AWSConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	VaultPath         string   `json:"vaultPath,omitempty"`
	AllowedAccountIDs []string `json:"allowedAccountIDs,omitempty"`
	AvailabiltyZones  []string `json:"availabilityZones,omitempty"`
	Region            string   `json:"region,omitempty"`
	KeyName           string   `json:"keyName,omitempty"` // ec2 key pair name
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type GCPConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Zones   []string `json:"zones,omitempty"`
	Region  string   `json:"region,omitempty"`
	Project string   `json:"project,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type Context struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Stacks []Stack `json:"stacks,omitempty"`

	Contact string `json:"contact,omitempty"`
	Project string `json:"project,omitempty"`

	BaseImage string `json:"baseImage,omitempty"`
}

type Stack struct {
	metav1.TypeMeta `json:",inline"`

	State      *StackState      `json:"state,omitempty"`
	Network    *StackNetwork    `json:"network,omitempty"`
	Tools      *StackTools      `json:"tools,omitempty"`
	Vault      *StackVault      `json:"vault,omitempty"`
	Kubernetes *StackKubernetes `json:"kubernetes,omitempty"`
	Custom     *StackCustom     `json:"custom,omitempty"`

	NodeGroups []NodeGroup `json:"nodeGroups,omitempty"`
}

type StackTools struct {
}

type StackNetwork struct {
	PeerContext string `json:"peerContext,omitempty"`
	NetworkCIDR string `json:"networkCIDR,omitempty"`
	PrivateZone string `json:"privateZone,omitempty"`
}

type StackVault struct {
}

type StackState struct {
	BucketPrefix string `json:"bucketPrefix,omitempty"`
	PublicZone   string `json:"publicZone,omitempty"`
}

type StackKubernetes struct {
}

type StackCustom struct {
	Name string `json:"name,omitempty"`
	Path string `json:"path,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type NodeGroup struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Role string `json:"role,omitempty"`

	Count int `json:"count,omitempty"`

	Volumes []Volume `json:"volumes,omitempty"`

	AWS *NodeGroupAWS `json:"aws,omitempty"`
}

type NodeGroupAWS struct {
	InstanceType string  `json:"instanceType,omitempty"`
	SpotPrice    float64 `json:"spotPrice,omitempty"`
}

type Volume struct {
	Name string `json:"name,omitempty"`
	Size int    `json:"size,omitempty"` // Size in GB

	AWS *VolumeAWS `json:"aws,omitempty"`
}

type VolumeAWS struct {
	Type string `json:"type,omitempty"` // gp2/st1
	// TODO: io1 (*needs more arguments) but would be good for at least etcd data dir
}
