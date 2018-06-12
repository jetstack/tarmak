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

	InstanceID   string `json:"instanceID,omitempty"`
	InstancePool string

	Spec   *InstanceSpec
	Status *InstanceStatus
}

// InstanceSpec defines the desired state of Instance
type InstanceSpec struct {
	PuppetManifestRef string
}

// InstanceStatus defines the observed state of Instance
type InstanceStatus struct {
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type InstanceList struct {
	metav1.TypeMeta
	metav1.ListMeta

	Items []Instance
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type PuppetTarget struct {
	metav1.TypeMeta
	metav1.ObjectMeta

	Source           ManifestSource
	Hash             string
	RequestTimestamp metav1.Time
}

type ManifestSource struct {
	S3 *S3ManifestSource
}

type S3ManifestSource struct {
	BucketName string
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type PuppetTargetList struct {
	metav1.TypeMeta
	metav1.ListMeta

	Items []PuppetTarget `json:"items"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type WingJob struct {
	metav1.TypeMeta
	metav1.ObjectMeta

	Spec   *WingJobSpec
	Status *WingJobStatus
}

type WingJobSpec struct {
	InstanceID       string
	Source           ManifestSource
	Operation        string
	RequestTimestamp metav1.Time
}

type WingJobStatus struct {
	Messages string
	ExitCode int
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type WingJobList struct {
	metav1.TypeMeta
	metav1.ListMeta

	Items []WingJob
}
