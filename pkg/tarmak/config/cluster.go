package config

import (
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	clusterv1alpha1 "github.com/jetstack/tarmak/pkg/apis/cluster/v1alpha1"
)

func newCluster(environment string, name string) *clusterv1alpha1.Cluster {
	c := &clusterv1alpha1.Cluster{}
	c.SetName(name)
	c.SetNamespace(environment)
	return c
}

// This creates a new cluster for a single cluster environment
func NewClusterSingle(environment string, name string) *clusterv1alpha1.Cluster {
	c := newCluster(environment, name)
	c.ServerPools = []clusterv1alpha1.ServerPool{
		*newServerPoolBastion(),
		*newServerPoolVault(),
		*newServerPoolEtcd(),
		*newServerPoolMaster(),
		*newServerPoolNode(),
	}
	return c
}

// This creates a new cluster for a multi cluster environment
func NewClusterMulti(environment string, name string) *clusterv1alpha1.Cluster {
	c := newCluster(environment, name)
	c.ServerPools = []clusterv1alpha1.ServerPool{
		*newServerPoolEtcd(),
		*newServerPoolMaster(),
		*newServerPoolNode(),
	}
	return c
}

// This creates a new hub for a multi cluster environment
func NewHub(environment string) *clusterv1alpha1.Cluster {
	c := newCluster(environment, "hub")
	c.ServerPools = []clusterv1alpha1.ServerPool{
		*newServerPoolBastion(),
		*newServerPoolVault(),
	}
	return c
}

func newServerPool() *clusterv1alpha1.ServerPool {
	sp := &clusterv1alpha1.ServerPool{}
	return sp
}

// This creates a bastion nodeGroup
func newServerPoolBastion() *clusterv1alpha1.ServerPool {
	sp := newServerPool()
	sp.Type = clusterv1alpha1.ServerPoolTypeBastion
	sp.MinCount = 1
	sp.MaxCount = 1
	sp.Size = clusterv1alpha1.ServerPoolSizeTiny
	sp.Volumes = []clusterv1alpha1.Volume{
		clusterv1alpha1.Volume{
			ObjectMeta: metav1.ObjectMeta{Name: "root"},
			Type:       clusterv1alpha1.VolumeTypeSSD,
		},
	}
	return sp
}

// This creates a etcd serverPool
func newServerPoolEtcd() *clusterv1alpha1.ServerPool {
	sp := newServerPool()
	sp.Type = clusterv1alpha1.ServerPoolTypeEtcd
	sp.MinCount = 3
	sp.MaxCount = 3
	sp.Size = clusterv1alpha1.ServerPoolSizeSmall
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

// This creates a vault serverPool
func newServerPoolVault() *clusterv1alpha1.ServerPool {
	sp := newServerPool()
	sp.Type = clusterv1alpha1.ServerPoolTypeVault
	sp.MinCount = 3
	sp.MaxCount = 3
	sp.Size = clusterv1alpha1.ServerPoolSizeTiny
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

// This creates a master serverPool
func newServerPoolMaster() *clusterv1alpha1.ServerPool {
	sp := newServerPool()
	sp.Type = clusterv1alpha1.ServerPoolTypeMaster
	sp.MinCount = 1
	sp.MaxCount = 1
	sp.Size = clusterv1alpha1.ServerPoolSizeMedium
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

// This creates a node serverPool
func newServerPoolNode() *clusterv1alpha1.ServerPool {
	sp := newServerPool()
	sp.Type = clusterv1alpha1.ServerPoolTypeNode
	sp.MinCount = 3
	sp.MaxCount = 3
	sp.Size = clusterv1alpha1.ServerPoolSizeMedium
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
