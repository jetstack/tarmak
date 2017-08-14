package node_group

import (
	"fmt"

	"github.com/Sirupsen/logrus"
	"github.com/hashicorp/go-multierror"

	"github.com/jetstack/tarmak/pkg/tarmak/config"
	"github.com/jetstack/tarmak/pkg/tarmak/interfaces"
	"github.com/jetstack/tarmak/pkg/tarmak/role"
)

var _ interfaces.NodeGroup = &NodeGroup{}

type NodeGroup struct {
	conf  *config.NodeGroup
	log   *logrus.Entry
	stack interfaces.Stack

	volumes []*Volume

	role *role.Role
}

func NewFromConfig(stack interfaces.Stack, conf *config.NodeGroup) (*NodeGroup, error) {
	nodeGroup := &NodeGroup{
		conf:  conf,
		stack: stack,
		log:   stack.Log().WithField("nodeGroup", conf.Name),
	}

	nodeGroup.role = stack.Role(conf.Role)
	if nodeGroup.role == nil {
		return nil, fmt.Errorf("role '%s' is not valid for this stack", conf.Role)
	}

	var result error

	for pos, _ := range conf.Volumes {
		volume, err := NewVolumeFromConfig(pos, &conf.Volumes[pos])
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
	return n.conf.Count
}

func (n *NodeGroup) AWSInstanceType() string {
	if n.conf.AWS != nil {
		return n.conf.AWS.InstanceType
	}
	return ""
}

func (n *NodeGroup) AWSSpotPrice() string {
	if n.conf.AWS != nil && n.conf.AWS.SpotPrice > 0 {
		return fmt.Sprintf("%f", n.conf.AWS.SpotPrice)
	}
	return ""
}
