// Copyright Jetstack Ltd. See LICENSE for details.
package v1alpha1

import (
	"time"

	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var zeroTime metav1.Time

func addDefaultingFuncs(scheme *runtime.Scheme) error {
	return RegisterDefaults(scheme)
}

func SetDefaults_Cluster(obj *Cluster) {
	// set creation time, if unset
	if obj.CreationTimestamp == zeroTime {
		obj.CreationTimestamp.Time = time.Now()
	}

	// set network object if nil {
	if obj.Network == nil {
		obj.Network = &Network{}
	}

	// set network.cidr if not existing
	if obj.Network.CIDR == "" {
		obj.Network.CIDR = "10.99.0.0/16"
	}

	// set kubernetes object if nil {
	if obj.Kubernetes == nil {
		obj.Kubernetes = &ClusterKubernetes{}
	}

	// set default kubernetes version
	if obj.Kubernetes.Version == "" {
		obj.Kubernetes.Version = "1.10.6"
	}

	// zone
	if obj.Kubernetes.Zone == "" {
		obj.Kubernetes.Zone = "cluster.local"
	}

	// podCIDR
	if obj.Kubernetes.PodCIDR == "" {
		obj.Kubernetes.PodCIDR = "100.64.0.0/16"
	}

	// serviceCIDR
	if obj.Kubernetes.ServiceCIDR == "" {
		obj.Kubernetes.ServiceCIDR = "10.254.0.0/16"
	}

	// clusterAutoscaler
	if obj.Kubernetes.ClusterAutoscaler == nil {
		obj.Kubernetes.ClusterAutoscaler = &ClusterKubernetesClusterAutoscaler{}
	}

	// tiller
	if obj.Kubernetes.Tiller == nil {
		obj.Kubernetes.Tiller = &ClusterKubernetesTiller{}
	}

	// dashboard
	if obj.Kubernetes.Dashboard == nil {
		obj.Kubernetes.Dashboard = &ClusterKubernetesDashboard{}
	}

	// logging
	if obj.LoggingSinks == nil {
		obj.LoggingSinks = []*LoggingSink{}
	}
	for _, loggingSink := range obj.LoggingSinks {
		if loggingSink.ElasticSearch != nil {
			if loggingSink.ElasticSearch.Host == "" {
				loggingSink.ElasticSearch.Host = "127.0.0.1"
			}
			if loggingSink.ElasticSearch.TLS == nil {
				loggingSink.ElasticSearch.TLS = boolPointer(true)
			}
			if loggingSink.ElasticSearch.Port == 0 {
				if *loggingSink.ElasticSearch.TLS {
					loggingSink.ElasticSearch.Port = 443
				} else {
					loggingSink.ElasticSearch.Port = 80
				}
			}
			if loggingSink.ElasticSearch.LogstashPrefix == "" {
				loggingSink.ElasticSearch.LogstashPrefix = "logstash"
			}
			if loggingSink.ElasticSearch.AmazonESProxy != nil {
				if loggingSink.ElasticSearch.AmazonESProxy.Port == 0 {
					loggingSink.ElasticSearch.AmazonESProxy.Port = allocateAmazonESProxyPort(obj.LoggingSinks)
				}
			}
		}

		if len(loggingSink.Types) == 0 {
			loggingSink.Types = []LoggingSinkType{"all"}
		}
	}

}

func boolPointer(x bool) *bool {
	return &x
}

func allocateAmazonESProxyPort(loggingSinks []*LoggingSink) int {

	allocatedPorts := make(map[int]struct{})
	for _, loggingSink := range loggingSinks {
		if loggingSink.ElasticSearch != nil {
			if loggingSink.ElasticSearch.AmazonESProxy != nil {
				allocatedPorts[loggingSink.ElasticSearch.AmazonESProxy.Port] = struct{}{}
			}
		}
	}

	currentPort := 9200
	for {
		if _, ok := allocatedPorts[currentPort]; ok {
			currentPort++
			continue
		}
		return currentPort
	}

}

func SetDefaults_Volume(obj *Volume) {
	// set creation time, if unset
	if obj.CreationTimestamp == zeroTime {
		obj.CreationTimestamp.Time = time.Now()
	}
}

func SetDefaults_InstancePool(obj *InstancePool) {
	// set creation time, if unset
	if obj.CreationTimestamp == zeroTime {
		obj.CreationTimestamp.Time = time.Now()
	}

	// set name to type, if unset
	if obj.Name == "" {
		obj.Name = obj.Type
	}

	// set image to default image
	if obj.Image == "" {
		obj.Image = "centos-puppet-agent"
	}

	// set a default size for volumes
	for pos, _ := range obj.Volumes {
		if obj.Volumes[pos].Size == nil {
			obj.Volumes[pos].Size = resource.NewQuantity(16*1024*1024*1024, resource.BinarySI)
		}
		if obj.Volumes[pos].Type == "" {
			obj.Volumes[pos].Type = VolumeTypeSSD
		}
	}
}
