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
		obj.Kubernetes = &Kubernetes{}
	}

	// set default kubernetes version
	if obj.Kubernetes.Version == "" {
		obj.Kubernetes.Version = "1.7.7"
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
		obj.Image = "centos-puppet-agent-latest-kernel"
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
