// Copyright Jetstack Ltd. See LICENSE for details.
package cluster

import (
	"errors"
	"fmt"
	"net"
	"path/filepath"

	"github.com/Sirupsen/logrus"
	"github.com/hashicorp/go-multierror"

	clusterv1alpha1 "github.com/jetstack/tarmak/pkg/apis/cluster/v1alpha1"
	tarmakv1alpha1 "github.com/jetstack/tarmak/pkg/apis/tarmak/v1alpha1"
	"github.com/jetstack/tarmak/pkg/tarmak/instance_pool"
	"github.com/jetstack/tarmak/pkg/tarmak/interfaces"
	"github.com/jetstack/tarmak/pkg/tarmak/role"
	"github.com/jetstack/tarmak/pkg/tarmak/stack"
	wingclient "github.com/jetstack/tarmak/pkg/wing/client"
)

// returns a server
type Cluster struct {
	conf *clusterv1alpha1.Cluster

	stacks []interfaces.Stack

	stackNetwork interfaces.Stack
	environment  interfaces.Environment
	networkCIDR  *net.IPNet
	log          *logrus.Entry

	wingClientset *wingclient.Clientset
	wingTunnel    interfaces.Tunnel

	imageIDs      map[string]string
	instancePools []interfaces.InstancePool
	roles         map[string]*role.Role
}

var _ interfaces.Cluster = &Cluster{}

func NewFromConfig(environment interfaces.Environment, conf *clusterv1alpha1.Cluster) (*Cluster, error) {
	cluster := &Cluster{
		conf:        conf,
		environment: environment,
		log:         environment.Log().WithField("cluster", conf.Name),
	}

	// validate server pools and setup stacks
	if err := cluster.validateInstancePools(); err != nil {
		return nil, err
	}

	cluster.roles = make(map[string]*role.Role)
	defineToolsRoles(cluster.roles)
	defineVaultRoles(cluster.roles)
	defineKubernetesRoles(cluster.roles)

	// setup instance pools
	var result error
	for pos, _ := range cluster.conf.InstancePools {
		instancePool := cluster.conf.InstancePools[pos]
		// create instance pools
		pool, err := instance_pool.NewFromConfig(cluster, &instancePool)
		if err != nil {
			result = multierror.Append(result, err)
			continue
		}
		cluster.instancePools = append(cluster.instancePools, pool)
	}

	return cluster, result
}

func (c *Cluster) InstancePools() []interfaces.InstancePool {
	return c.instancePools
}

func (c *Cluster) ListHosts() ([]interfaces.Host, error) {
	return c.Environment().Provider().ListHosts()
}

func (c *Cluster) InstancePoolsMap() (instancePoolsMap map[string][]*clusterv1alpha1.InstancePool) {
	instancePoolsMap = make(map[string][]*clusterv1alpha1.InstancePool)
	for pos, _ := range c.conf.InstancePools {
		pool := &c.conf.InstancePools[pos]
		_, ok := instancePoolsMap[pool.Type]
		if !ok {
			instancePoolsMap[pool.Type] = []*clusterv1alpha1.InstancePool{pool}
		} else {
			instancePoolsMap[pool.Type] = append(instancePoolsMap[pool.Type], pool)
		}
	}
	return instancePoolsMap
}

// validate hub instancePool types
func validateHubTypes(poolMap map[string][]*clusterv1alpha1.InstancePool, clusterType string) (result error) {
	if len(poolMap[clusterv1alpha1.InstancePoolTypeBastion]) != 1 {
		result = multierror.Append(result, fmt.Errorf("a hub needs to have exactly one '%s' server pool", clusterv1alpha1.InstancePoolTypeBastion))
	}

	if len(poolMap[clusterv1alpha1.InstancePoolTypeVault]) != 1 {
		result = multierror.Append(result, fmt.Errorf("a hub needs to have exactly one '%s' server pool", clusterv1alpha1.InstancePoolTypeVault))
	}

	return result
}

// validate cluster instancePool types
func validateClusterTypes(poolMap map[string][]*clusterv1alpha1.InstancePool, clusterType string) (result error) {
	if len(poolMap[clusterv1alpha1.InstancePoolTypeEtcd]) != 1 {
		result = multierror.Append(result, fmt.Errorf("a %s needs to have exactly one '%s' server pool", clusterType, clusterv1alpha1.InstancePoolTypeEtcd))
	}

	if len(poolMap[clusterv1alpha1.InstancePoolTypeMaster]) < 1 {
		result = multierror.Append(result, fmt.Errorf("a %s needs to have more than one '%s' server pool", clusterType, clusterv1alpha1.InstancePoolTypeMaster))
	}

	return result
}

// validate server pools
func (c *Cluster) validateInstancePools() (result error) {
	poolMap := c.InstancePoolsMap()
	clusterType := c.Type()
	allowedTypes := make(map[string]bool)
	c.stacks = []interfaces.Stack{}

	// Validate hub for cluster-single and hub
	if clusterType == clusterv1alpha1.ClusterTypeClusterSingle || clusterType == clusterv1alpha1.ClusterTypeHub {
		err := validateHubTypes(poolMap, clusterType)
		if err != nil {
			result = multierror.Append(result, err)
		}
		allowedTypes[clusterv1alpha1.InstancePoolTypeJenkins] = true
		allowedTypes[clusterv1alpha1.InstancePoolTypeBastion] = true
		allowedTypes[clusterv1alpha1.InstancePoolTypeVault] = true

		if s, err := stack.New(c, tarmakv1alpha1.StackNameState); err != nil {
			result = multierror.Append(result, err)
		} else {
			c.stacks = append(c.stacks, s)
		}

		// make the choice between deploying into existing VPC or creating a new one
		if _, ok := c.Config().Network.ObjectMeta.Annotations["tarmak.io/existing-vpc-id"]; ok {
			if s, err := stack.New(c, tarmakv1alpha1.StackNameExistingNetwork); err != nil {
				result = multierror.Append(result, err)
			} else {
				c.stacks = append(c.stacks, s)
			}
		} else {
			if s, err := stack.New(c, tarmakv1alpha1.StackNameNetwork); err != nil {
				result = multierror.Append(result, err)
			} else {
				c.stacks = append(c.stacks, s)
			}
		}

		if s, err := stack.New(c, tarmakv1alpha1.StackNameTools); err != nil {
			result = multierror.Append(result, err)
		} else {
			c.stacks = append(c.stacks, s)
		}

		if s, err := stack.New(c, tarmakv1alpha1.StackNameVault); err != nil {
			result = multierror.Append(result, err)
		} else {
			c.stacks = append(c.stacks, s)
		}
	}

	// validate cluster for cluster-*
	if clusterType == clusterv1alpha1.ClusterTypeClusterSingle || clusterType == clusterv1alpha1.ClusterTypeClusterMulti {
		err := validateClusterTypes(poolMap, clusterType)
		if err != nil {
			result = multierror.Append(result, err)
		}
		allowedTypes[clusterv1alpha1.InstancePoolTypeEtcd] = true
		allowedTypes[clusterv1alpha1.InstancePoolTypeMaster] = true
		allowedTypes[clusterv1alpha1.InstancePoolTypeWorker] = true

		if s, err := stack.New(c, tarmakv1alpha1.StackNameKubernetes); err != nil {
			result = multierror.Append(result, err)
		} else {
			c.stacks = append(c.stacks, s)
		}
	}

	// check for unsupported pool types
	for poolType := range poolMap {
		if _, ok := allowedTypes[poolType]; ok {
			continue
		}
		result = multierror.Append(result, fmt.Errorf("the pool type %s is not supported for a %s", poolType, clusterType))
	}

	return result
}

// Determine if this Cluster is a cluster or hub, single or multi environment
func (c *Cluster) Type() string {
	if c.conf.Type != "" {
		return c.conf.Type
	}

	if len(c.Environment().Tarmak().Config().Clusters(c.Environment().Name())) == 1 {
		return clusterv1alpha1.ClusterTypeClusterSingle
	}
	if c.Name() == clusterv1alpha1.ClusterTypeHub {
		return clusterv1alpha1.ClusterTypeHub
	}
	return clusterv1alpha1.ClusterTypeClusterMulti
}

func (c *Cluster) RemoteState(stackName string) string {
	// special case for the existing network stack, allows other stacks to not
	// care which is deployed
	if stackName == "network-existing-vpc" {
		stackName = "network"
	}
	return c.Environment().Provider().RemoteState(c.Environment().Name(), c.Name(), stackName)
}

func (c *Cluster) Region() string {
	return c.conf.Location
}

func (c *Cluster) Subnets() (subnets []clusterv1alpha1.Subnet) {
	zones := make(map[string]bool)

	for _, sp := range c.conf.InstancePools {
		for _, subnet := range sp.Subnets {
			zones[subnet.Zone] = true
		}
	}

	for zone, _ := range zones {
		subnets = append(subnets, clusterv1alpha1.Subnet{Zone: zone})
	}

	return subnets
}

// This methods aggregates all images of the pools
func (c *Cluster) Images() []string {
	images := make(map[string]bool)
	for _, sp := range c.conf.InstancePools {
		images[sp.Image] = true
	}

	imagesDistinct := []string{}
	for image, _ := range images {
		imagesDistinct = append(imagesDistinct, image)
	}

	return imagesDistinct
}

func (c *Cluster) ImageIDs() (map[string]string, error) {
	if c.imageIDs == nil {
		imageMap, err := c.Environment().Tarmak().Packer().IDs()
		if err != nil {
			return nil, err
		}
		c.imageIDs = imageMap
	}

	return c.imageIDs, nil
}

func (c *Cluster) getNetworkCIDR() (*net.IPNet, error) {
	if c.stackNetwork == nil {
		return nil, errors.New("no network stack found")
	}

	netIntf, ok := c.stackNetwork.Variables()["network"]
	if !ok {
		return nil, errors.New("no network variable in stack network found")
	}

	net, ok := netIntf.(*net.IPNet)
	if !ok {
		return nil, errors.New("network variable has unexpected typ")
	}

	return net, nil
}

func (c *Cluster) NetworkCIDR() *net.IPNet {
	return c.networkCIDR
}

func (c *Cluster) APITunnel() interfaces.Tunnel {
	return c.Environment().Tarmak().SSH().Tunnel(
		"bastion",
		fmt.Sprintf("api.%s.%s", c.ClusterName(), c.Environment().Config().PrivateZone),
		6443,
	)
}

func (c *Cluster) Validate() error {
	return nil
}

func (c *Cluster) Stacks() []interfaces.Stack {
	return c.stacks
}

func (c *Cluster) Stack(name string) interfaces.Stack {
	for _, stack := range c.stacks {
		if stack.Name() == name {
			return stack
		}
	}
	return nil
}

func (c *Cluster) Environment() interfaces.Environment {
	return c.environment
}

func (c *Cluster) ClusterName() string {
	return fmt.Sprintf("%s-%s", c.environment.Name(), c.conf.Name)
}

func (c *Cluster) Name() string {
	return c.conf.Name
}

func (c *Cluster) Config() *clusterv1alpha1.Cluster {
	return c.conf.DeepCopy()
}

func (c *Cluster) ConfigPath() string {
	return filepath.Join(c.Environment().Tarmak().ConfigPath(), c.ClusterName())
}

func (c *Cluster) SSHConfigPath() string {
	return filepath.Join(c.ConfigPath(), "ssh_config")
}

func (c *Cluster) SSHHostKeysPath() string {
	return filepath.Join(c.ConfigPath(), "ssh_known_hosts")
}

func (c *Cluster) Log() *logrus.Entry {
	return c.log
}

func (c *Cluster) Role(roleName string) *role.Role {
	if c.roles != nil {
		if role, ok := c.roles[roleName]; ok {
			return role
		}
	}
	return nil
}

func (c *Cluster) Roles() (roles []*role.Role) {
	roleMap := map[string]bool{}
	for _, instancePool := range c.InstancePools() {
		r := instancePool.Role()
		if _, ok := roleMap[r.Name()]; !ok {
			roles = append(roles, r)
			roleMap[r.Name()] = true
		}
	}
	return roles
}

func (c *Cluster) Parameters() map[string]string {
	return map[string]string{
		"name":        c.Name(),
		"environment": c.Environment().Name(),
		"provider":    c.Environment().Provider().String(),
	}
}

func (c *Cluster) Variables() map[string]interface{} {
	output := c.environment.Variables()

	// TODO: refactor me
	/*
		if c.conf.Contact != "" {
			output["contact"] = c.conf.Contact
		}
		if c.conf.Project != "" {
			output["project"] = c.conf.Project
		}

		if c.imageID != nil {
			output["centos_ami"] = map[string]string{
				c.environment.Provider().Region(): *c.imageID,
			}
		}
	*/

	output["name"] = c.Name()

	return output
}
