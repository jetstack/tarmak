// Copyright Jetstack Ltd. See LICENSE for details.
package instance_pool

import (
	"errors"
	"fmt"
	"sort"

	"github.com/Sirupsen/logrus"
	"github.com/hashicorp/go-multierror"

	clusterv1alpha1 "github.com/jetstack/tarmak/pkg/apis/cluster/v1alpha1"
	"github.com/jetstack/tarmak/pkg/tarmak/interfaces"
	"github.com/jetstack/tarmak/pkg/tarmak/role"
	"github.com/jetstack/tarmak/pkg/tarmak/utils"
)

var _ interfaces.InstancePool = &InstancePool{}

type InstancePool struct {
	conf *clusterv1alpha1.InstancePool
	log  *logrus.Entry

	cluster interfaces.Cluster

	volumes    []*Volume
	rootVolume *Volume

	instanceType string

	role *role.Role
}

func NewFromConfig(cluster interfaces.Cluster, conf *clusterv1alpha1.InstancePool) (*InstancePool, error) {
	instancePool := &InstancePool{
		conf:    conf,
		cluster: cluster,
		log:     cluster.Log().WithField("instancePool", conf.Name),
	}

	instancePool.role = cluster.Role(conf.Type)
	if instancePool.role == nil {
		return nil, fmt.Errorf("role '%s' is not valid for this cluster", conf.Type)
	}

	// validate instance size with cloud provider
	provider := cluster.Environment().Provider()
	instanceType, err := provider.InstanceType(conf.Size)
	if err != nil {
		return nil, fmt.Errorf("instanceType '%s' is not valid for this provier", conf.Size)
	}
	instancePool.instanceType = instanceType

	var result error

	count := 0
	for pos, _ := range conf.Volumes {
		volume, err := NewVolumeFromConfig(count, provider, &conf.Volumes[pos])
		if err != nil {
			result = multierror.Append(result, err)
			continue
		}
		if volume.Name() == "root" {
			instancePool.rootVolume = volume
		} else {
			count++
			instancePool.volumes = append(instancePool.volumes, volume)
		}
	}

	if instancePool.rootVolume == nil {
		return nil, errors.New("no root volume given")
	}

	return instancePool, result
}

func (n *InstancePool) Role() *role.Role {
	return n.role
}

func (n *InstancePool) Image() string {
	return n.conf.Image
}

//Get unique list of zones of instance pool
func (n *InstancePool) Zones() (zones []string) {
	for _, subnet := range n.Config().Subnets {
		zones = append(zones, subnet.Zone)
	}

	zones = utils.RemoveDuplicateStrings(zones)
	sort.Strings(zones)

	return zones
}

func (n *InstancePool) Name() string {
	if n.conf.Name == "" {
		return n.Role().Name()
	}
	return n.conf.Name
}

func (n *InstancePool) Config() *clusterv1alpha1.InstancePool {
	return n.conf.DeepCopy()
}

// This returns a DNS compatible name
func (n *InstancePool) DNSName() string {
	return n.Role().Prefix("-") + n.Name()
}

// This returns a TF compatible name
func (n *InstancePool) TFName() string {
	return n.Role().Prefix("_") + n.Name()
}

func (n *InstancePool) Volumes() (volumes []interfaces.Volume) {
	for _, volume := range n.volumes {
		volumes = append(volumes, volume)
	}
	return volumes
}

func (n *InstancePool) RootVolume() interfaces.Volume {
	return n.rootVolume
}

func (n *InstancePool) MinCount() int {
	return n.conf.MinCount
}

func (n *InstancePool) MaxCount() int {
	return n.conf.MaxCount
}

func (n *InstancePool) InstanceType() string {
	return n.instanceType
}

func (n *InstancePool) SpotPrice() string {
	return n.conf.SpotPrice
}
