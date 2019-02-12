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
		obj.Kubernetes.Version = "1.11.5"
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

	if obj.Kubernetes.Calico == nil {
		obj.Kubernetes.Calico = &ClusterKubernetesCalico{
			Backend:     "etcd",
			EnableTypha: false,
		}
	}

	if obj.Kubernetes.Calico.Backend == "" {
		obj.Kubernetes.Calico.Backend = "etcd"
	}

	if obj.Kubernetes.Calico.EnableTypha &&
		obj.Kubernetes.Calico.TyphaReplicas == nil {
		obj.Kubernetes.Calico.TyphaReplicas = intPointer(1)
	}

	if obj.Kubernetes.Heapster == nil {
		obj.Kubernetes.Heapster = &ClusterKubernetesHeapster{
			Enabled: true,
		}
	}

	if obj.Kubernetes.Grafana == nil {
		obj.Kubernetes.Grafana = &ClusterKubernetesGrafana{
			Enabled: true,
		}
	}

	if obj.Kubernetes.InfluxDB == nil {
		obj.Kubernetes.InfluxDB = &ClusterKubernetesInfluxDB{
			Enabled: true,
		}
	}

	// EBS encryption off if Amazon interface used
	// but EBSEncrypted not specified
	if obj.Amazon == nil {
		obj.Amazon = &ClusterAmazon{}
	}
	if obj.Amazon.EBSEncrypted == nil {
		obj.Amazon.EBSEncrypted = boolPointer(false)
	}
	if obj.Amazon.AdditionalIAMPolicies == nil {
		obj.Amazon.AdditionalIAMPolicies = []string{}
	}

	// logging
	if obj.LoggingSinks == nil {
		obj.LoggingSinks = []*LoggingSink{}
	}
	for _, loggingSink := range obj.LoggingSinks {
		if loggingSink.Elasticsearch != nil {
			if loggingSink.Elasticsearch.Host == "" {
				loggingSink.Elasticsearch.Host = "127.0.0.1"
			}
			if loggingSink.Elasticsearch.TLS == nil {
				loggingSink.Elasticsearch.TLS = boolPointer(true)
			}
			if loggingSink.Elasticsearch.Port == 0 {
				if *loggingSink.Elasticsearch.TLS {
					loggingSink.Elasticsearch.Port = 443
				} else {
					loggingSink.Elasticsearch.Port = 80
				}
			}
			if loggingSink.Elasticsearch.LogstashPrefix == "" {
				loggingSink.Elasticsearch.LogstashPrefix = "logstash"
			}
			if loggingSink.Elasticsearch.AmazonESProxy != nil {
				if loggingSink.Elasticsearch.AmazonESProxy.Port == 0 {
					loggingSink.Elasticsearch.AmazonESProxy.Port = allocateAmazonESProxyPort(obj.LoggingSinks)
				}
			}
		}

		if len(loggingSink.Types) == 0 {
			loggingSink.Types = []LoggingSinkType{"all"}
		}
	}

	// Vault
	if obj.VaultHelper == nil {
		obj.VaultHelper = new(ClusterVaultHelper)
	}

}

func boolPointer(x bool) *bool {
	return &x
}

func floatPointer(x float64) *float64 {
	return &x
}

func intPointer(x int) *int {
	return &x
}

func allocateAmazonESProxyPort(loggingSinks []*LoggingSink) int {

	allocatedPorts := make(map[int]struct{})
	for _, loggingSink := range loggingSinks {
		if loggingSink.Elasticsearch != nil {
			if loggingSink.Elasticsearch.AmazonESProxy != nil {
				allocatedPorts[loggingSink.Elasticsearch.AmazonESProxy.Port] = struct{}{}
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
		if obj.Type == InstancePoolTypeWorker {
			obj.Image = ImageBaseDefaultWorker
		} else {
			obj.Image = ImageBaseDefault
		}
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

	if obj.Amazon == nil {
		obj.Amazon = new(InstancePoolAmazon)
	}
	if obj.Amazon.AdditionalIAMPolicies == nil {
		obj.Amazon.AdditionalIAMPolicies = []string{}
	}
}

func SetDefaults_ClusterKubernetesAPIServerAmazonAccessLogs(obj *ClusterKubernetesAPIServerAmazonAccessLogs) {
	if obj.Enabled == nil {
		if len(obj.Bucket) > 0 {
			obj.Enabled = boolPointer(true)
		} else {
			obj.Enabled = boolPointer(false)
		}
	}

	if obj.Interval == nil {
		in := 5
		obj.Interval = &in
	}
}

func SetDefaults_ClusterKubernetesClusterAutoscaler(obj *ClusterKubernetesClusterAutoscaler) {
	if obj.ScaleDownUtilizationThreshold == nil {
		obj.ScaleDownUtilizationThreshold = floatPointer(0.5)
	}
}
