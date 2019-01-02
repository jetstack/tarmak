// Copyright Jetstack Ltd. See LICENSE for details.
package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/jetstack/tarmak/pkg/apis/wing/common"
)

// MachineSetSpec defines the desired state of MachineSet
type MachineSetSpec struct {
	// Replicas is the number of desired replicas.
	// This is a pointer to distinguish between explicit zero and unspecified.
	// Defaults to 1.
	// +optional
	MinReplicas *int32 `json:"minReplicas,omitempty"`
	MaxReplicas *int32 `json:"maxReplicas,omitempty"`

	// MinReadySeconds is the minimum number of seconds for which a newly created machine should be ready.
	// Defaults to 0 (machine will be considered available as soon as it is ready)
	// +optional
	MinReadySeconds int32 `json:"minReadySeconds,omitempty"`

	// Selector is a label query over machines that should match the replica count.
	// Label keys and values that must match in order to be controlled by this MachineSet.
	// It must match the machine template's labels.
	// More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/#label-selectors
	Selector metav1.LabelSelector `json:"selector"`

	// Template is the object that describes the machine that will be created if
	// insufficient replicas are detected.
	// +optional
	Template MachineTemplateSpec `json:"template,omitempty"`
}

// MachineTemplateSpec describes the data needed to create a Machine from a template
type MachineTemplateSpec struct {
	// Standard object's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Specification of the desired behavior of the machine.
	// More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#spec-and-status
	Spec MachineSpec `json:"spec,omitempty"`
}

// MachineSetStatus defines the observed state of MachineSet
type MachineSetStatus struct {
	// Replicas is the most recently observed number of replicas.
	Replicas int32 `json:"replicas"`

	// The number of replicas that have labels matching the labels of the machine template of the MachineSet.
	// +optional
	FullyLabeledReplicas int32 `json:"fullyLabeledReplicas,omitempty"`

	// The number of ready replicas for this MachineSet. A machine is considered ready when the node has been created and is "Ready".
	// +optional
	ReadyReplicas int32 `json:"readyReplicas,omitempty"`

	// The number of available replicas (ready for at least minReadySeconds) for this MachineSet.
	// +optional
	AvailableReplicas int32 `json:"availableReplicas,omitempty"`

	// ObservedGeneration reflects the generation of the most recently observed MachineSet.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// In the event that there is a terminal problem reconciling the
	// replicas, both ErrorReason and ErrorMessage will be set. ErrorReason
	// will be populated with a succinct value suitable for machine
	// interpretation, while ErrorMessage will contain a more verbose
	// string suitable for logging and human consumption.
	//
	// These fields should not be set for transitive errors that a
	// controller faces that are expected to be fixed automatically over
	// time (like service outages), but instead indicate that something is
	// fundamentally wrong with the MachineTemplates's spec or the configuration of
	// the machine controller, and that manual intervention is required. Examples
	// of terminal errors would be invalid combinations of settings in the
	// spec, values that are unsupported by the machine controller, or the
	// responsible machine controller itself being critically misconfigured.
	//
	// Any transient errors that occur during the reconciliation of Machines
	// can be added as events to the MachineSet object and/or logged in the
	// controller's output.
	// +optional
	ErrorReason common.MachineStatusError `json:"errorReason,omitempty"`
	//ErrorReason *string `json:"errorReason,omitempty"`
	// +optional
	ErrorMessage *string `json:"errorMessage,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:openapi-gen=true
// +resource:path=wing

// MachineSet is the Schema for the machinesets API
// +k8s:openapi-gen=true
type MachineSet struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   *MachineSetSpec   `json:"spec,omitempty"`
	Status *MachineSetStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MachineSetList contains a list of MachineSet
type MachineSetList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MachineSet `json:"items"`
}
