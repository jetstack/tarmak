// Copyright Jetstack Ltd. See LICENSE for details.
// Copyright © 2017 The Kubicorn Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	InstancePoolSizeTiny   = "tiny"
	InstancePoolSizeSmall  = "small"
	InstancePoolSizeMedium = "medium"
	InstancePoolSizeLarge  = "large"
)

const (
	InstancePoolTypeMaster     = "master"
	InstancePoolTypeWorker     = "worker"
	InstancePoolTypeEtcd       = "etcd"
	InstancePoolTypeBastion    = "bastion" // bastion node with public IP
	InstancePoolTypeJenkins    = "jenkins" // jenkins CI/CD node
	InstancePoolTypeVault      = "vault"
	InstancePoolTypeAll        = "all"         // master + worker + etcd
	InstancePoolTypeMasterEtcd = "master-etcd" // master + etcd
	InstancePoolTypeHybrid     = "hybrid"      // master + worker
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type InstancePool struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Identifier        string                  `json:"identifier,omitempty"`
	MinCount          int                     `json:"minCount,omitempty"`
	MaxCount          int                     `json:"maxCount,omitempty"`
	Type              string                  `json:"type,omitempty"`
	Image             string                  `json:"image,omitempty"`
	Size              string                  `json:"size,omitempty"`
	SpotPrice         string                  `json:"spotPrice,omitempty"`
	BootstrapScripts  []string                `json:"bootstrapScripts,omitempty"`
	Subnets           []*Subnet               `json:"subnets,omitempty"`
	Firewalls         []*Firewall             `json:"firewalls,omitempty"`
	Volumes           []Volume                `json:"volumes,omitempty"`
	Kubernetes        *InstancePoolKubernetes `json:"kubernetes,omitempty"`
	AllowCIDRs        []string                `json:"allowCIDRs, omitempty"`

	// Amazon specific settings for that instance pool
	Amazon *InstancePoolAmazon `json:"amazon,omitempty"`
}

type InstancePoolKubernetes struct {
	Version string `json:"version,omitempty"`
}

// Amazon specific settings for that instance pool
type InstancePoolAmazon struct {
	// This fields contains ARNs for additional IAM policies to be added to
	// this instance pool
	AdditionalIAMPolicies []string `json:"additionalIAMPolicies,omitempty"`
}
