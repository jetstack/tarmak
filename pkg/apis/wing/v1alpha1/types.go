// Copyright Jetstack Ltd. See LICENSE for details.
package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ApplyOperation  = "apply"
	DryRunOperation = "dry-run"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type Instance struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	InstanceID   string `json:"instanceID,omitempty"`
	InstancePool string `json:"instancePool,omitempty"`

	Spec   *InstanceSpec   `json:"spec,omitempty"`
	Status *InstanceStatus `json:"status,omitempty"`
}

// InstanceSpec defines the desired state of Instance
type InstanceSpec struct {
	PuppetTargetRef string `json:"puppetTargetRef"`
}

// InstanceStatus defines the observed state of Instance
type InstanceStatus struct {
	PuppetTargetRef string `json:"puppetTargetRef"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type InstanceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []Instance `json:"items"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type PuppetTarget struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Source ManifestSource `json:"source"`
	Hash   string         `json:"hash,omitempty"` // hash of manifests, prefixed with type (eg: sha256:xyz)
}

type ManifestSource struct {
	S3   *S3ManifestSource   `json:"s3"`
	File *FileManifestSource `json:"file"`
}

type S3ManifestSource struct {
	BucketName string `json:"bucketName"`
	Path       string `json:"path"`
}

type FileManifestSource struct {
	Path string `json:"path"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type PuppetTargetList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []PuppetTarget `json:"items"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type WingJob struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   *WingJobSpec   `json:"spec,omitempty"`
	Status *WingJobStatus `json:"status,omitempty"`
}

type WingJobSpec struct {
	InstanceName     string      `json:"instanceName,omitempty"`
	PuppetTargetRef  string      `json:"puppetTargetRef"`
	Operation        string      `json:"operation"`
	RequestTimestamp metav1.Time `json:"requestTimestamp,omitempty"`
}

type WingJobStatus struct {
	Messages            string      `json:"messages,omitempty"`
	ExitCode            int         `json:"exitCode,omitempty"`
	LastUpdateTimestamp metav1.Time `json:"lastUpdateTimestamp,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type WingJobList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []WingJob `json:"items"`
}
