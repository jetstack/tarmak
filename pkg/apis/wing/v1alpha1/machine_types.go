// Copyright Jetstack Ltd. See LICENSE for details.
package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/jetstack/tarmak/pkg/apis/wing/common"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type Machine struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   *MachineSpec   `json:"spec,omitempty"`
	Status *MachineStatus `json:"status,omitempty"`
}

// MachineSpec defines the desired state of Machine
type MachineSpec struct {
	Converge *MachineSpecManifest `json:"converge,omitempty"`
	DryRun   *MachineSpecManifest `json:"dryRun,omitempty"`
}

//  InstaceSpecManifest defines location and hash for a specific manifest
type MachineSpecManifest struct {
	Path             string      `json:"path,omitempty"`             // PATH to manifests (tar.gz)
	Hash             string      `json:"hash,omitempty"`             // hash of manifests, prefixed with type (eg: sha256:xyz)
	RequestTimestamp metav1.Time `json:"requestTimestamp,omitempty"` // timestamp when a converge was requested
}

// MachineStatus defines the observed state of Machine
type MachineStatus struct {
	Converge *MachineStatusManifest `json:"converge,omitempty"`
	DryRun   *MachineStatusManifest `json:"dryRun,omitempty"`
}

//  InstaceSpecManifest defines the state and hash of a run manifest
type MachineStatusManifest struct {
	State               common.MachineManifestState `json:"state,omitempty"`
	Hash                string                      `json:"hash,omitempty"`                // hash of manifests, prefixed with type (eg: sha256:xyz)
	LastUpdateTimestamp metav1.Time                 `json:"lastUpdateTimestamp,omitempty"` // timestamp when a converge was requested
	Messages            []string                    `json:"messages,omitempty"`            // contains output of the retries
	ExitCodes           []int                       `json:"exitCodes,omitempty"`           // return code of the retries
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type MachineList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []Machine `json:"items"`
}
