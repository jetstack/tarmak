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
	CloudAmazon       = "amazon"
	CloudAzure        = "azure"
	CloudGoogle       = "google"
	CloudBaremetal    = "baremetal"
	CloudDigitalOcean = "digitalocean"
)

const (
	ClusterTypeHub           = "hub"
	ClusterTypeClusterSingle = "cluster-single"
	ClusterTypeClusterMulti  = "cluster-multi"
)

const (
	// represents Terraform in a destroy state
	StateDestroy = "destroy"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +resource:path=clusters

type Cluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	CloudId           string         `json:"cloudId,omitempty"`
	InstancePools     []InstancePool `json:"instancePools,omitempty"`
	Cloud             string         `json:"cloud,omitempty"`
	Location          string         `json:"location,omitempty"`
	Network           *Network       `json:"network,omitempty"`
	Values            *Values        `json:"values,omitempty"`
	KubernetesAPI     *KubernetesAPI `json:"kubernetesAPI,omitempty"`
	GroupIdentifier   string         `json:"groupIdentifier,omitempty"`

	Environment string             `json:"environment,omitempty"`
	Kubernetes  *ClusterKubernetes `json:"kubernetes,omitempty"`

	Type string `json:"-"` // This specifies if a cluster is a hub, single or multi
}

type ClusterKubernetes struct {
	Zone              string                              `json:"zone,omitempty"`
	Version           string                              `json:"version,omitempty"`
	PodCIDR           string                              `json:"podCIDR,omitempty"`
	ServiceCIDR       string                              `json:"serviceCIDR,omitempty"`
	ClusterAutoscaler *ClusterKubernetesClusterAutoscaler `json:"clusterAutoscaler,omitempty"`
	Tiller            *ClusterKubernetesTiller            `json:"tiller,omitempty"`
	Dashboard         *ClusterKubernetesDashboard         `json:"dashboard,omitempty"`
}

type ClusterKubernetesClusterAutoscaler struct {
	Enabled bool   `json:"enabled,omitempty"`
	Image   string `json:"image,omitempty"`
	Version string `json:"version,omitempty"`
}

type ClusterKubernetesTiller struct {
	Enabled bool   `json:"enabled,omitempty"`
	Image   string `json:"image,omitempty"`
	Version string `json:"version,omitempty"`
}

type ClusterKubernetesDashboard struct {
	Enabled bool   `json:"enabled,omitempty"`
	Image   string `json:"image,omitempty"`
	Version string `json:"version,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type ClusterList struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Items []Cluster `json:"items"`
}

func NewCluster(name string) *Cluster {
	return &Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
}
