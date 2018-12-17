// Copyright Jetstack Ltd. See LICENSE for details.
package wing

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/jetstack/tarmak/pkg/apis/wing/common"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type Machine struct {
	metav1.TypeMeta
	metav1.ObjectMeta

	Spec   *MachineSpec
	Status *MachineStatus
}

// MachineSpec defines the desired state of Machine
type MachineSpec struct {
	Converge *MachineSpecManifest
	DryRun   *MachineSpecManifest
}

//  InstaceSpecManifest defines location and hash for a specific manifest
type MachineSpecManifest struct {
	Path             string
	Hash             string
	RequestTimestamp metav1.Time
}

// MachineStatus defines the observed state of Machine
type MachineStatus struct {
	Converge *MachineStatusManifest
	DryRun   *MachineStatusManifest
}

//  InstaceSpecManifest defines the state and hash of a run manifest
type MachineStatusManifest struct {
	State               common.MachineManifestState
	Hash                string
	LastUpdateTimestamp metav1.Time
	Messages            []string
	ExitCodes           []int
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type MachineList struct {
	metav1.TypeMeta
	// +optional
	metav1.ListMeta

	Items []Machine
}
