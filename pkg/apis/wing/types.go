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
	PuppetTargetRef string
}

// InstanceStatus defines the observed state of Instance
type InstanceStatus struct {
	PuppetTargetRef string
	Converged       bool
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

	Source ManifestSource
	Hash   string
}

type ManifestSource struct {
	S3   *S3ManifestSource
	File *FileManifestSource
}

type S3ManifestSource struct {
	BucketName string
	Path       string
}

type FileManifestSource struct {
	Path string
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
	InstanceName     string
	PuppetTargetRef  string
	Operation        string
	RequestTimestamp metav1.Time
}

type WingJobStatus struct {
	Messages            string
	ExitCode            int
	Completed           bool
	LastUpdateTimestamp metav1.Time
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type WingJobList struct {
	metav1.TypeMeta
	metav1.ListMeta

	Items []WingJob
}
