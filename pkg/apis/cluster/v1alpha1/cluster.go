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

const (
	PrometheusModeFull                      = "Full"
	PrometheusModeExternalExportersOnly     = "ExternalExportersOnly"
	PrometheusModeExternalScrapeTargetsOnly = "ExternalScrapeTargetsOnly"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:openapi-gen=true
// +resource:path=clusters

type Cluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	CloudId           string         `json:"cloudId,omitempty"`
	InstancePools     []InstancePool `json:"instancePools,omitempty"`
	Cloud             string         `json:"cloud,omitempty"`
	Location          string         `json:"location,omitempty"`
	Network           *Network       `json:"network,omitempty"`
	LoggingSinks      []*LoggingSink `json:"loggingSinks,omitempty"`
	Values            *Values        `json:"values,omitempty"`
	KubernetesAPI     *KubernetesAPI `json:"kubernetesAPI,omitempty"`
	GroupIdentifier   string         `json:"groupIdentifier,omitempty"`

	Environment string             `json:"environment,omitempty"`
	Kubernetes  *ClusterKubernetes `json:"kubernetes,omitempty"`

	Type string `json:"-"` // This specifies if a cluster is a hub, single or multi

	// Amazon specific options
	Amazon *ClusterAmazon `json:"amazon"`
}

// ClusterAmazon offers Amazon-specific settings for that instance pool
type ClusterAmazon struct {
	// This fields contains ARNs for additional IAM policies to be added to
	// this instance pool
	AdditionalIAMPolicies []string `json:"additionalIAMPolicies,omitempty"`
	// When set to true, AWS Elastic Block Storage volumes are encrypted
	EBSEncrypted *bool `json:"ebsEncrypted,omitempty"`
}

type ClusterKubernetes struct {
	Zone              string                              `json:"zone,omitempty"`
	Version           string                              `json:"version,omitempty"`
	PodCIDR           string                              `json:"podCIDR,omitempty"`
	ServiceCIDR       string                              `json:"serviceCIDR,omitempty"`
	ClusterAutoscaler *ClusterKubernetesClusterAutoscaler `json:"clusterAutoscaler,omitempty"`
	Tiller            *ClusterKubernetesTiller            `json:"tiller,omitempty"`
	Dashboard         *ClusterKubernetesDashboard         `json:"dashboard,omitempty"`
	APIServer         *ClusterKubernetesAPIServer         `json:"apiServer,omitempty"`
	PodSecurityPolicy *ClusterPodSecurityPolicy           `json:"podSecurityPolicy,omitempty"`
	Prometheus        *ClusterKubernetesPrometheus        `json:"prometheus,omitempty"`
}

type ClusterKubernetesClusterAutoscaler struct {
	Enabled                       bool                                                `json:"enabled,omitempty"`
	Image                         string                                              `json:"image,omitempty"`
	Version                       string                                              `json:"version,omitempty"`
	ScaleDownUtilizationThreshold *float64                                            `json:"scaleDownUtilizationThreshold,omitempty"`
	Overprovisioning              *ClusterKubernetesClusterAutoscalerOverprovisioning `json:"overprovisioning,omitempty"`
}

type ClusterKubernetesClusterAutoscalerOverprovisioning struct {
	Enabled                      bool   `json:"enabled,omitempty"`
	Image                        string `json:"image,omitempty"`
	Version                      string `json:"version,omitempty"`
	ReservedMillicoresPerReplica int    `json:"reservedMillicoresPerReplica,omitempty"`
	ReservedMegabytesPerReplica  int    `json:"reservedMegabytesPerReplica,omitempty"`
	CoresPerReplica              int    `json:"coresPerReplica,omitempty"`
	NodesPerReplica              int    `json:"nodesPerReplica,omitempty"`
	ReplicaCount                 int    `json:"replicaCount,omitempty"`
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

type ClusterKubernetesAPIServer struct {
	// expose the API server through a public load balancer
	Public     bool     `json:"public,omitempty"`
	AllowCIDRs []string `json:"allowCIDRs,omitempty"`

	EnableAdmissionControllers  []string `json:"enableAdmissionControllers,omitempty"`
	DisableAdmissionControllers []string `json:"disableAdmissionControllers,omitempty"`

	// OIDC
	OIDC *ClusterKubernetesAPIServerOIDC `json:"oidc,omitempty"`

	// AWS specifc options
	Amazon *ClusterKubernetesAPIServerAmazon `json:"amazon,omitempty"`
}

type ClusterKubernetesAPIServerOIDC struct {
	// The client ID for the OpenID Connect client, must be set if oidc-issuer-url is set.
	ClientID string `json:"clientID,omitempty" hiera:"kubernetes::apiserver::oidc_client_id"`

	// If provided, the name of a custom OpenID Connect claim for specifying
	// user groups. The claim value is expected to be a string or array of
	// strings. This flag is experimental, please see the authentication
	// documentation for further details.
	GroupsClaim string `json:"groupsClaim,omitempty" hiera:"kubernetes::apiserver::oidc_groups_claim"`

	// If provided, all groups will be prefixed with this value to prevent
	// conflicts with other authentication strategies.
	GroupsPrefix string `json:"groupsPrefix,omitempty" hiera:"kubernetes::apiserver::oidc_groups_prefix"`
	// The URL of the OpenID issuer, only HTTPS scheme will be accepted. If
	// set, it will be used to verify the OIDC JSON Web Token (JWT).
	IssuerURL string `json:"issuerURL,omitempty" hiera:"kubernetes::apiserver::oidc_issuer_url"`

	// Comma-separated list of allowed JOSE asymmetric signing algorithms. JWTs
	// with a 'alg' header value not in this list will be rejected. Values are
	// defined by RFC 7518 https://tools.ietf.org/html/rfc7518#section-3.1.
	// (default [RS256])
	SigningAlgs []string `json:"signingAlgs,omitempty" hiera:"kubernetes::apiserver::oidc_signing_algs"`

	// The OpenID claim to use as the user name. Note that claims other than
	// the default ('sub') is not guaranteed to be unique and immutable. This
	// flag is experimental, please see the authentication documentation for
	// further details. (default "sub")
	UsernameClaim string `json:"usernameClaim,omitempty" hiera:"kubernetes::apiserver::oidc_username_claim"`

	// If provided, all usernames will be prefixed with this value. If not
	// provided, username claims other than 'email' are prefixed by the issuer
	// URL to avoid clashes. To skip any prefixing, provide the value '-'.
	UsernamePrefix string `json:"usernamePrefix,omitempty" hiera:"kubernetes::apiserver::oidc_username_prefix"`
}

type ClusterKubernetesAPIServerAmazon struct {
	PublicELBAccessLogs   *ClusterKubernetesAPIServerAmazonAccessLogs `json:"publicELBAccessLogs,omitempty"`
	InternalELBAccessLogs *ClusterKubernetesAPIServerAmazonAccessLogs `json:"internalELBAccessLogs,omitempty"`
}

type ClusterKubernetesAPIServerAmazonAccessLogs struct {
	Enabled      *bool  `json:"enabled,omitempty"`
	Bucket       string `json:"bucket,omitempty"`
	BucketPrefix string `json:"bucketPrefix,omitempty"`
	Interval     *int   `json:"interval,omitempty"`
}

type ClusterPodSecurityPolicy struct {
	Enabled bool `json:"enabled,omitempty"`
}

// Configure the cluster internal deployment of prometheus
type ClusterKubernetesPrometheus struct {
	// Enable a cluster internal prometheus deployment, default: true
	Enabled bool `json:"enabled,omitempty"`
	// Mode defines which components are installed
	Mode string `json:"mode,omitempty"`
}

// +k8s:openapi-gen=true
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
