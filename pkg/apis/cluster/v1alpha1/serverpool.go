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
	ServerPoolTypeMaster     = "master"
	ServerPoolTypeNode       = "node"
	ServerPoolTypeEtcd       = "etcd"
	ServerPoolTypeBastion    = "bastion" // bastion node with public IP
	ServerPoolTypeVault      = "vault"
	ServerPoolTypeAll        = "all"         // master + node + etcd
	ServerPoolTypeMasterEtcd = "master-etcd" // master + etcd
	ServerPoolTypeHybrid     = "hybrid"      // master + node
)

type ServerPool struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Identifier        string      `json:"identifier,omitempty"`
	MinCount          int         `json:"minCount,omitempty"`
	MaxCount          int         `json:"maxCount,omitempty"`
	Type              string      `json:"type,omitempty"`
	Name              string      `json:"name,omitempty"`
	Image             string      `json:"image,omitempty"`
	Size              string      `json:"size,omitempty"`
	SpotPrice         string      `json:"spotPrice,omitempty"`
	BootstrapScripts  []string    `json:"bootstrapScripts,omitempty"`
	Subnets           []*Subnet   `json:"subnets,omitempty"`
	Firewalls         []*Firewall `json:"firewalls,omitempty"`
	Volumes           []*Volume   `json:"volumes,omitempty"`
}
