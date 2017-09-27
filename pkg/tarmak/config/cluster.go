package config

import (
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	clusterv1alpha1 "github.com/jetstack/tarmak/pkg/apis/cluster/v1alpha1"
)

func newCluster(environment string, name string) *clusterv1alpha1.Cluster {
	c := &clusterv1alpha1.Cluster{}
	c.Name = name
	c.Environment = environment
	return c
}

// This creates a new cluster for a single cluster environment
func NewClusterSingle(environment string, name string) *clusterv1alpha1.Cluster {
	c := newCluster(environment, name)
	c.Type = clusterv1alpha1.ClusterTypeClusterSingle
	c.InstancePools = []clusterv1alpha1.InstancePool{
		*newInstancePoolBastion(),
		*newInstancePoolVault(),
		*newInstancePoolEtcd(),
		*newInstancePoolMaster(),
		*newInstancePoolWorker(),
	}
	ApplyDefaults(c)
	return c
}

// This creates a new cluster for a multi cluster environment
func NewClusterMulti(environment string, name string) *clusterv1alpha1.Cluster {
	c := newCluster(environment, name)
	c.Type = clusterv1alpha1.ClusterTypeClusterMulti
	c.InstancePools = []clusterv1alpha1.InstancePool{
		*newInstancePoolEtcd(),
		*newInstancePoolMaster(),
		*newInstancePoolWorker(),
	}
	ApplyDefaults(c)
	return c
}

// This creates a new hub for a multi cluster environment
func NewHub(environment string) *clusterv1alpha1.Cluster {
	c := newCluster(environment, clusterv1alpha1.ClusterTypeHub)
	c.Type = clusterv1alpha1.ClusterTypeHub
	c.InstancePools = []clusterv1alpha1.InstancePool{
		*newInstancePoolBastion(),
		*newInstancePoolVault(),
	}
	ApplyDefaults(c)
	return c
}

func newInstancePool() *clusterv1alpha1.InstancePool {
	sp := &clusterv1alpha1.InstancePool{}
	return sp
}

// This creates a bastion instancePool
func newInstancePoolBastion() *clusterv1alpha1.InstancePool {
	sp := newInstancePool()
	sp.Type = clusterv1alpha1.InstancePoolTypeBastion
	sp.MinCount = 1
	sp.MaxCount = 1
	sp.Size = clusterv1alpha1.InstancePoolSizeTiny
	sp.Volumes = []clusterv1alpha1.Volume{
		clusterv1alpha1.Volume{
			ObjectMeta: metav1.ObjectMeta{Name: "root"},
			Type:       clusterv1alpha1.VolumeTypeSSD,
		},
	}
	return sp
}

// This creates a etcd instancePool
func newInstancePoolEtcd() *clusterv1alpha1.InstancePool {
	sp := newInstancePool()
	sp.Type = clusterv1alpha1.InstancePoolTypeEtcd
	sp.MinCount = 3
	sp.MaxCount = 3
	sp.Size = clusterv1alpha1.InstancePoolSizeSmall
	sp.Volumes = []clusterv1alpha1.Volume{
		clusterv1alpha1.Volume{
			ObjectMeta: metav1.ObjectMeta{Name: "root"},
			Type:       clusterv1alpha1.VolumeTypeSSD,
		},
		clusterv1alpha1.Volume{
			ObjectMeta: metav1.ObjectMeta{Name: "data"},
			Size:       resource.NewQuantity(5*1024*1024*1024, resource.BinarySI),
			Type:       clusterv1alpha1.VolumeTypeSSD,
		},
	}
	return sp
}

// This creates a vault instancePool
func newInstancePoolVault() *clusterv1alpha1.InstancePool {
	sp := newInstancePool()
	sp.Type = clusterv1alpha1.InstancePoolTypeVault
	sp.MinCount = 3
	sp.MaxCount = 3
	sp.Size = clusterv1alpha1.InstancePoolSizeTiny
	sp.Volumes = []clusterv1alpha1.Volume{
		clusterv1alpha1.Volume{
			ObjectMeta: metav1.ObjectMeta{Name: "root"},
			Type:       clusterv1alpha1.VolumeTypeSSD,
		},
		clusterv1alpha1.Volume{
			ObjectMeta: metav1.ObjectMeta{Name: "data"},
			Size:       resource.NewQuantity(5*1024*1024*1024, resource.BinarySI),
			Type:       clusterv1alpha1.VolumeTypeSSD,
		},
	}
	return sp
}

// This creates a master instancePool
func newInstancePoolMaster() *clusterv1alpha1.InstancePool {
	sp := newInstancePool()
	sp.Type = clusterv1alpha1.InstancePoolTypeMaster
	sp.MinCount = 1
	sp.MaxCount = 1
	sp.Size = clusterv1alpha1.InstancePoolSizeMedium
	sp.Volumes = []clusterv1alpha1.Volume{
		clusterv1alpha1.Volume{
			ObjectMeta: metav1.ObjectMeta{Name: "root"},
			Type:       clusterv1alpha1.VolumeTypeSSD,
		},
		clusterv1alpha1.Volume{
			ObjectMeta: metav1.ObjectMeta{Name: "docker"},
			Size:       resource.NewQuantity(10*1024*1024*1024, resource.BinarySI),
			Type:       clusterv1alpha1.VolumeTypeSSD,
		},
	}
	return sp
}

// This creates a node instancePool
func newInstancePoolWorker() *clusterv1alpha1.InstancePool {
	sp := newInstancePool()
	sp.Type = clusterv1alpha1.InstancePoolTypeWorker
	sp.MinCount = 3
	sp.MaxCount = 3
	sp.Size = clusterv1alpha1.InstancePoolSizeMedium
	sp.Volumes = []clusterv1alpha1.Volume{
		clusterv1alpha1.Volume{
			ObjectMeta: metav1.ObjectMeta{Name: "root"},
			Type:       clusterv1alpha1.VolumeTypeSSD,
		},
		clusterv1alpha1.Volume{
			ObjectMeta: metav1.ObjectMeta{Name: "docker"},
			Size:       resource.NewQuantity(50*1024*1024*1024, resource.BinarySI),
			Type:       clusterv1alpha1.VolumeTypeSSD,
		},
	}
	return sp
}
