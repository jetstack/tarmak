package v1alpha1

import (
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient=true
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +resource:path=contexts

type Context struct {
	metav1.TypeMeta
	metav1.ObjectMeta

	Spec   ContextSpec
	Status ContextStatus
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ContextList is a list of Contexts
type ContextList struct {
	metav1.TypeMeta
	metav1.ListMeta

	BaseParams

	Items []Context
}

type ContextSpec struct {
	NodeGroups []NodeGroup
	Version    string
}

type ContextStatus struct {
	Fine    bool
	Version string
}

type NodeGroup struct {
	metav1.TypeMeta
	metav1.ObjectMeta

	RootDiskSize resource.Quantity

	NodeGroupDisk []NodeGroupDisk

	Role string // vault/etcd/master/worker/all
}

type NodeGroupDisk struct {
	Name     string
	DiskSize resource.Quantity
}

// this represents single instances
type NodeGroupTypeOrderedInstances struct {
	Count int
}

type NodeGroupTypeScalingGroup struct {
	MinCount     int
	MaxCount     int
	DesiredCount int // 0 == do not change set count
}

type NodeGroupAWS struct {
	AdditionalIAMPolicy string
	SpotPrice           float64
	InstanceType        string
}

// +genclient=true
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +resource:path=environments

// Environment is at exactly one provider and region and could contain multiple (cluster-) contexts
type Environment struct {
	metav1.TypeMeta
	metav1.ObjectMeta

	BaseParams

	Spec   ContextSpec
	Status ContextStatus
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// EnvironmentList is a list of Environments
type EnvironmentList struct {
	metav1.TypeMeta
	metav1.ListMeta

	Items []Environment
}

type BaseParams struct {
	Contact        string
	Project        string
	TrustedIPRange []string
}
