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

const (
	LoggingSinkTypePlatform    = LoggingSinkType("platform")    // kernel, systemd, platform namespaces
	LoggingSinkTypeApplication = LoggingSinkType("application") // all other namespaces
	LoggingSinkTypeAudit       = LoggingSinkType("audit")       // api server audit logs
	LoggingSinkTypeAll         = LoggingSinkType("all")
)

type LoggingSinkType string

type LoggingSink struct {
	ElasticSearch *LoggingSinkElasticSearch `json:"elasticsearch,omitempty"`
	Types         []LoggingSinkType         `json:"types,omitempty"`
}

type LoggingSinkElasticSearch struct {
	// https://fluentbit.io/documentation/0.12/output/elasticsearch.html
	Host       string      `json:"host,omitempty"`
	Port       int         `json:"port,omitempty"`
	TLS        bool        `json:"tls,omitempty"`
	TLSVerify  bool        `json:"tlsVerify,omitempty"`
	AWSESProxy *AWSESProxy `json:"awsESProxy,omitempty"`
}

type AWSESProxy struct {
	Port int `json:"port,omitempty"`
}
