// Copyright Jetstack Ltd. See LICENSE for details.
package wing

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type Instance struct {
	metav1.TypeMeta
	metav1.ObjectMeta

	InstanceID   string
	InstancePool string

	Spec   *InstanceSpec
	Status *InstanceStatus
}

// InstanceSpec defines the desired state of Instance
type InstanceSpec struct {
	Converge *InstanceSpecManifest
	DryRun   *InstanceSpecManifest
}

//  InstaceSpecManifest defines location and hash for a specific manifest
type InstanceSpecManifest struct {
	Path             string
	Hash             string
	RequestTimestamp metav1.Time
}

// InstanceStatus defines the observed state of Instance
type InstanceStatus struct {
	Converge *InstanceStatusManifest
	DryRun   *InstanceStatusManifest
}

//  InstaceSpecManifest defines the state and hash of a run manifest
type InstanceManifestState string
type InstanceStatusManifest struct {
	State               InstanceManifestState
	Hash                string
	LastUpdateTimestamp metav1.Time
	Messages            []string
	ExitCodes           []int
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type InstanceList struct {
	metav1.TypeMeta
	// +optional
	metav1.ListMeta

	Items []Instance
}
