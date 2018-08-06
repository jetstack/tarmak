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

import "testing"

func TestLoggingDefaults(t *testing.T) {

	cluster := &Cluster{
		LoggingSinks: []*LoggingSink{
			&LoggingSink{
				ElasticSearch: &LoggingSinkElasticSearch{
					TLS: boolPointer(true),
				},
			},
			&LoggingSink{
				ElasticSearch: &LoggingSinkElasticSearch{
					TLS: boolPointer(false),
				},
			},
			&LoggingSink{
				ElasticSearch: &LoggingSinkElasticSearch{},
			},
		},
	}

	SetDefaults_Cluster(cluster)

	if cluster.LoggingSinks == nil {
		t.Errorf("logging sinks not set")
	} else {
		for index, loggingSink := range cluster.LoggingSinks {
			if loggingSink.ElasticSearch == nil {
				t.Errorf("elasticsearch is not set for logging sink %d", index)
			} else {
				if loggingSink.ElasticSearch.TLS == nil {
					t.Errorf("elasticsearch tls is not set for logging sink %d", index)
				} else {
					if (index == 0 || index == 2) && *loggingSink.ElasticSearch.TLS != true {
						t.Errorf("elasticsearch for logging sink %d does not have TLS enabled", index)
					}
					if index == 1 && *loggingSink.ElasticSearch.TLS != false {
						t.Errorf("elasticsearch for logging sink %d has TLS enabled", index)
					}
				}
			}
		}
	}
}
