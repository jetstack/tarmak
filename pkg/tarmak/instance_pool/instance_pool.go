// Copyright Jetstack Ltd. See LICENSE for details.
package instance_pool

import (
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/hashicorp/go-multierror"
	"github.com/sirupsen/logrus"
	validation "k8s.io/apimachinery/pkg/apis/meta/v1/validation"
	"k8s.io/apimachinery/pkg/util/validation/field"

	"net"

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
		return nil, fmt.Errorf("instanceType '%s' is not valid for this provider: %v", conf.Size, err)
	}
	instancePool.instanceType = instanceType

	// validate minCount <= maxCount or minCount == maxCount if role is stateful
	// if only one of the two values are set, we should default to the other
	if instancePool.Config().MinCount == 0 && instancePool.Config().MaxCount == 0 {
		return nil, errors.New("minCount and maxCount both not set or set to 0")
	}

	if instancePool.Config().MinCount == 0 {
		instancePool.conf.MinCount = instancePool.conf.MaxCount
	} else if instancePool.Config().MaxCount == 0 {
		instancePool.conf.MaxCount = instancePool.conf.MinCount
	}

	if instancePool.Config().MinCount > instancePool.Config().MaxCount {
		return nil, fmt.Errorf("minCount is larger than maxCount. minCount=%d maxCount=%d", instancePool.Config().MinCount, instancePool.Config().MaxCount)
	}

	if instancePool.Role().Stateful && instancePool.Config().MinCount != instancePool.Config().MaxCount {
		return nil, fmt.Errorf("minCount does not equal maxCount but role is stateful. minCount=%d maxCount=%d", instancePool.Config().MinCount, instancePool.Config().MaxCount)
	}

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

func (n *InstancePool) AmazonAdditionalIAMPolicies() string {
	policies := []string{}

	// add cluster wide policies
	if a := n.cluster.Config().Amazon; a != nil {
		policies = append(policies, a.AdditionalIAMPolicies...)
	}

	// add instance template specfic policies
	if a := n.Config().Amazon; a != nil {
		policies = append(policies, a.AdditionalIAMPolicies...)
	}

	// TODO: check for duplicates here

	for pos, _ := range policies {
		policies[pos] = fmt.Sprintf(`"%s"`, policies[pos])
	}

	return fmt.Sprintf("[%s]", strings.Join(policies, ","))
}

func (n *InstancePool) Validate() (result error) {
	return n.ValidateAllowCIDRs()
}

func (n *InstancePool) ValidateAllowCIDRs() (result error) {
	for _, cidr := range n.Config().AllowCIDRs {
		_, _, err := net.ParseCIDR(cidr)
		if err != nil {
			result = multierror.Append(result, fmt.Errorf("%s is not a valid CIDR format", cidr))
		}
	}

	return result
}

func (n *InstancePool) Labels() (string, error) {
	labelMap := make(map[string]string)
	for _, label := range n.conf.Labels {
		labelMap[label.Key] = label.Value
	}

	err := validation.ValidateLabels(labelMap, &field.Path{})
	if len(err) != 0 {
		return "", fmt.Errorf("%v", err)
	}

	var labels []string
	for _, label := range n.conf.Labels {
		labels = append(labels, fmt.Sprintf("  %s: \"%s\"", label.Key, label.Value))
	}

	return strings.Join(labels, "\n"), nil
}

func (n *InstancePool) Taints() (string, error) {
	var taints []string
	var result error

	err := n.validTaints()
	if err != nil {
		return "", err
	}

	for _, taint := range n.conf.Taints {
		taints = append(taints, fmt.Sprintf("  %s: \"%s:%s\"", taint.Key, taint.Value, taint.Effect))
	}

	return strings.Join(taints, "\n"), result
}

func (n *InstancePool) validTaints() error {
	var result error

	validKey := regexp.MustCompile(`^[a-zA-Z0-9][\w_\-\.]*\/?[\w_\-\.]*[a-zA-Z0-9]$`)
	validValue := regexp.MustCompile(`^[a-zA-Z0-9][\w_\-\.]*[a-zA-Z0-9]$`)
	validEffect := regexp.MustCompile(`^PreferNoSchedule|NoSchedule|NoExecute$`)

	for _, taint := range n.conf.Taints {
		if !validKey.MatchString(taint.Key) {
			result = multierror.Append(result, fmt.Errorf("key was invalid for taint: %+v", taint))
		}
		if len(taint.Value) > 0 && !validValue.MatchString(taint.Value) {
			result = multierror.Append(result, fmt.Errorf("value was invalid for taint: %+v", taint))
		}
		if !validEffect.MatchString(taint.Effect) {
			result = multierror.Append(result, fmt.Errorf("effect was invalid for taint: %+v", taint))
		}
	}

	return result
}
