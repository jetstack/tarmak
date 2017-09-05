package context

import (
	"errors"
	"fmt"
	"net"
	"path/filepath"

	"github.com/Sirupsen/logrus"
	"github.com/hashicorp/go-multierror"

	clusterv1alpha1 "github.com/jetstack/tarmak/pkg/apis/cluster/v1alpha1"
	"github.com/jetstack/tarmak/pkg/tarmak/interfaces"
	"github.com/jetstack/tarmak/pkg/tarmak/stack"
)

const (
	ContextTypeHub           = "hub"
	ContextTypeClusterSingle = "cluster-single"
	ContextTypeClusterMulti  = "cluster-multi"
)

// returns a server
type Context struct {
	conf *clusterv1alpha1.Cluster

	stacks []interfaces.Stack

	stackNetwork interfaces.Stack
	environment  interfaces.Environment
	imageID      *string
	networkCIDR  *net.IPNet
	log          *logrus.Entry
}

var _ interfaces.Context = &Context{}

func NewFromConfig(environment interfaces.Environment, conf *clusterv1alpha1.Cluster) (*Context, error) {
	context := &Context{
		conf:        conf,
		environment: environment,
		log:         environment.Log().WithField("context", conf.Name),
	}

	// validate server pools and setup stacks
	if err := context.validateServerPools(); err != nil {
		return nil, err
	}

	return context, nil
}

func (c *Context) ServerPools() []clusterv1alpha1.ServerPool {
	return c.conf.ServerPools
}

func (c *Context) ServerPoolsMap() (serverPoolsMap map[string][]*clusterv1alpha1.ServerPool) {
	serverPoolsMap = make(map[string][]*clusterv1alpha1.ServerPool)
	for pos, _ := range c.conf.ServerPools {
		pool := &c.conf.ServerPools[pos]
		_, ok := serverPoolsMap[pool.Type]
		if !ok {
			serverPoolsMap[pool.Type] = []*clusterv1alpha1.ServerPool{pool}
		} else {
			serverPoolsMap[pool.Type] = append(serverPoolsMap[pool.Type], pool)
		}
	}
	return serverPoolsMap
}

// validate hub serverPool types
func validateHubTypes(poolMap map[string][]*clusterv1alpha1.ServerPool, clusterType string) (result error) {
	if len(poolMap[clusterv1alpha1.ServerPoolTypeBastion]) != 1 {
		result = multierror.Append(result, fmt.Errorf("a hub needs to have exactly one '%s' server pool", clusterv1alpha1.ServerPoolTypeBastion))
	}

	if len(poolMap[clusterv1alpha1.ServerPoolTypeVault]) != 1 {
		result = multierror.Append(result, fmt.Errorf("a hub needs to have exactly one '%s' server pool", clusterv1alpha1.ServerPoolTypeVault))
	}

	return result
}

// validate cluster serverPool types
func validateClusterTypes(poolMap map[string][]*clusterv1alpha1.ServerPool, clusterType string) (result error) {
	if len(poolMap[clusterv1alpha1.ServerPoolTypeEtcd]) != 1 {
		result = multierror.Append(result, fmt.Errorf("a %s needs to have exactly one '%s' server pool", clusterType, clusterv1alpha1.ServerPoolTypeEtcd))
	}

	if len(poolMap[clusterv1alpha1.ServerPoolTypeMaster]) < 1 {
		result = multierror.Append(result, fmt.Errorf("a %s needs to have more than one '%s' server pool", clusterType, clusterv1alpha1.ServerPoolTypeMaster))
	}

	return result
}

// validate server pools
func (c *Context) validateServerPools() (result error) {
	poolMap := c.ServerPoolsMap()
	clusterType := c.Type()
	allowedTypes := make(map[string]bool)
	c.stacks = []interfaces.Stack{}

	// Validate hub for cluster-single and hub
	if clusterType == ContextTypeClusterSingle || clusterType == ContextTypeHub {
		err := validateHubTypes(poolMap, clusterType)
		if err != nil {
			result = multierror.Append(result, err)
		}
		allowedTypes[clusterv1alpha1.ServerPoolTypeJenkins] = true
		allowedTypes[clusterv1alpha1.ServerPoolTypeBastion] = true
		allowedTypes[clusterv1alpha1.ServerPoolTypeVault] = true

		if s, err := stack.New(c, stack.StackNameState); err != nil {
			result = multierror.Append(result, err)
		} else {
			c.stacks = append(c.stacks, s)
		}

		if s, err := stack.New(c, stack.StackNameNetwork); err != nil {
			result = multierror.Append(result, err)
		} else {
			c.stacks = append(c.stacks, s)
		}

		if s, err := stack.New(c, stack.StackNameTools); err != nil {
			result = multierror.Append(result, err)
		} else {
			c.stacks = append(c.stacks, s)
		}

		if s, err := stack.New(c, stack.StackNameVault); err != nil {
			result = multierror.Append(result, err)
		} else {
			c.stacks = append(c.stacks, s)
		}
	}

	// validate cluster for cluster-*
	if clusterType == ContextTypeClusterSingle || clusterType == ContextTypeClusterMulti {
		err := validateClusterTypes(poolMap, clusterType)
		if err != nil {
			result = multierror.Append(result, err)
		}
		allowedTypes[clusterv1alpha1.ServerPoolTypeEtcd] = true
		allowedTypes[clusterv1alpha1.ServerPoolTypeMaster] = true
		allowedTypes[clusterv1alpha1.ServerPoolTypeNode] = true

		if s, err := stack.New(c, stack.StackNameKubernetes); err != nil {
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

// Determine if this Context is a cluster or hub, single or multi environment
func (c *Context) Type() string {
	if len(c.Environment().Contexts()) == 1 {
		return ContextTypeClusterSingle
	}
	if c.Name() == ContextTypeHub {
		return ContextTypeHub
	}
	return ContextTypeClusterMulti
}

func (c *Context) RemoteState(stackName string) string {
	return c.Environment().Provider().RemoteState(c.Environment().Name(), c.Name(), stackName)
}

func (c *Context) Region() string {
	return c.conf.Location
}

func (c *Context) Subnets() (subnets []clusterv1alpha1.Subnet) {
	zones := make(map[string]bool)

	for _, sp := range c.conf.ServerPools {
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
func (c *Context) Images() []string {
	images := make(map[string]bool)
	for _, sp := range c.conf.ServerPools {
		images[sp.Image] = true
	}

	imagesDistinct := []string{}
	for image, _ := range images {
		imagesDistinct = append(imagesDistinct, image)
	}

	return imagesDistinct
}

func (c *Context) getNetworkCIDR() (*net.IPNet, error) {
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

func (c *Context) NetworkCIDR() *net.IPNet {
	return c.networkCIDR
}

func (c *Context) APITunnel() interfaces.Tunnel {
	return c.Environment().Tarmak().SSH().Tunnel(
		"bastion",
		fmt.Sprintf("api.%s.%s", c.ContextName(), c.stackNetwork.Variables()["private_zone"]),
		6443,
	)

}

func (c *Context) SetImageID(imageID string) {
	c.imageID = &imageID
}

func (c *Context) Validate() error {
	return nil
}

func (c *Context) Stacks() []interfaces.Stack {
	return c.stacks
}

func (c *Context) Environment() interfaces.Environment {
	return c.environment
}

func (c *Context) ContextName() string {
	return fmt.Sprintf("%s-%s", c.environment.Name(), c.conf.Name)
}

func (c *Context) Name() string {
	return c.conf.Name
}

func (c *Context) ConfigPath() string {
	return filepath.Join(c.Environment().Tarmak().ConfigPath(), c.ContextName())
}

func (c *Context) SSHConfigPath() string {
	return filepath.Join(c.ConfigPath(), "ssh_config")
}

func (c *Context) SSHHostKeysPath() string {
	return filepath.Join(c.ConfigPath(), "ssh_known_hosts")
}

func (c *Context) Log() *logrus.Entry {
	return c.log
}

func (c *Context) Variables() map[string]interface{} {
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
