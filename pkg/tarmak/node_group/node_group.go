package node_group

import (
	"fmt"

	"github.com/Sirupsen/logrus"
	"github.com/hashicorp/go-multierror"

	clusterv1alpha1 "github.com/jetstack/tarmak/pkg/apis/cluster/v1alpha1"
	"github.com/jetstack/tarmak/pkg/tarmak/interfaces"
	"github.com/jetstack/tarmak/pkg/tarmak/role"
)

var _ interfaces.NodeGroup = &NodeGroup{}

type NodeGroup struct {
	conf  *clusterv1alpha1.ServerPool
	log   *logrus.Entry
	stack interfaces.Stack

	volumes []*Volume

	instanceType string

	role *role.Role
}

func NewFromConfig(stack interfaces.Stack, conf *clusterv1alpha1.ServerPool) (*NodeGroup, error) {
	nodeGroup := &NodeGroup{
		conf:  conf,
		stack: stack,
		log:   stack.Log().WithField("nodeGroup", conf.Name),
	}

	nodeGroup.role = stack.Role(conf.Type)
	if nodeGroup.role == nil {
		return nil, fmt.Errorf("role '%s' is not valid for this stack", conf.Type)
	}

	// validate instance size with cloud provider
	provider := stack.Context().Environment().Provider()
	instanceType, err := provider.InstanceType(conf.Size)
	if err != nil {
		return nil, fmt.Errorf("instanceType '%s' is not valid for this provier", conf.Size)
	}
	nodeGroup.instanceType = instanceType

	var result error

	for pos, _ := range conf.Volumes {
		volume, err := NewVolumeFromConfig(pos, provider, &conf.Volumes[pos])
		if err != nil {
			result = multierror.Append(result, err)
			continue
		}
		nodeGroup.volumes = append(nodeGroup.volumes, volume)
	}

	return nodeGroup, result
}

func (n *NodeGroup) Role() *role.Role {
	return n.role
}

func (n *NodeGroup) Name() string {
	if n.conf.Name == "" {
		return n.Role().Name()
	}
	return n.conf.Name
}

// This returns a DNS compatible name
func (n *NodeGroup) DNSName() string {
	return n.Role().Prefix("-") + n.Name()
}

// This returns a TF compatible name
func (n *NodeGroup) TFName() string {
	return n.Role().Prefix("_") + n.Name()
}

func (n *NodeGroup) Volumes() (volumes []interfaces.Volume) {
	for _, volume := range n.volumes {
		volumes = append(volumes, volume)
	}
	return volumes
}

func (n *NodeGroup) Count() int {
	// TODO: this needs to be replaced by Max/Min
	return n.conf.MaxCount
}

func (n *NodeGroup) InstanceType() string {
	return n.instanceType
}

func (n *NodeGroup) SpotPrice() string {
	return n.conf.SpotPrice
}
