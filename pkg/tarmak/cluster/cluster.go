// Copyright Jetstack Ltd. See LICENSE for details.
package cluster

import (
	"fmt"
	"net"
	"path/filepath"

	"github.com/hashicorp/go-multierror"
	"github.com/sirupsen/logrus"

	clusterv1alpha1 "github.com/jetstack/tarmak/pkg/apis/cluster/v1alpha1"
	"github.com/jetstack/tarmak/pkg/tarmak/instance_pool"
	"github.com/jetstack/tarmak/pkg/tarmak/interfaces"
	"github.com/jetstack/tarmak/pkg/tarmak/role"
	wingclient "github.com/jetstack/tarmak/pkg/wing/client"
)

// returns a server
type Cluster struct {
	conf *clusterv1alpha1.Cluster

	environment interfaces.Environment
	networkCIDR *net.IPNet
	log         *logrus.Entry

	wingClientset *wingclient.Clientset
	wingTunnel    interfaces.Tunnel

	imageIDs      map[string]string
	instancePools []interfaces.InstancePool
	roles         map[string]*role.Role
	// state records the state of Terraform to determine
	// whether we are destroying or applying. This allows other
	// components of Tarmak to make better decisions
	state string
}

var _ interfaces.Cluster = &Cluster{}

func NewFromConfig(environment interfaces.Environment, conf *clusterv1alpha1.Cluster) (*Cluster, error) {
	cluster := &Cluster{
		conf:        conf,
		environment: environment,
		log:         environment.Log().WithField("cluster", conf.Name),
	}

	// validate instance pools
	if err := cluster.validateInstancePools(); err != nil {
		return nil, err
	}

	// validate network setup
	if err := cluster.validateNetwork(); err != nil {
		return nil, err
	}

	//validate logging
	if err := cluster.validateLoggingSinks(); err != nil {
		return nil, err
	}

	cluster.roles = make(map[string]*role.Role)
	defineToolsRoles(cluster.roles)
	defineVaultRoles(cluster.roles)
	defineKubernetesRoles(cluster.roles)

	// populate role information if the API server should be public
	if k := cluster.Config().Kubernetes; k != nil {
		if apiServer := k.APIServer; apiServer != nil && apiServer.Public == true {
			if master := cluster.Role("master"); master != nil {
				master.AWS.ELBAPIPublic = true
			}
		}
	}

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

func (c *Cluster) InstancePool(roleName string) interfaces.InstancePool {
	for _, instancePool := range c.instancePools {
		if instancePool.Role().Name() == roleName {
			return instancePool
		}
	}
	return nil
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
	return nil
	//return fmt.Errorf("refactore me!")
}

/*
	//poolMap := c.InstancePoolsMap()
	clusterType := c.Type()
	//allowedTypes := make(map[string]bool)
	c.stacks = []interfaces.Stack{}

	// only and always add kubernetes stack
	/*if s, err := stack.New(c, tarmakv1alpha1.StackNameKubernetes); err != nil {
		result = multierror.Append(result, err)
	} else {
		c.stacks = append(c.stacks, s)
	}

	return result

	// Validate hub for cluster-single and hub
	if true || clusterType == clusterv1alpha1.ClusterTypeClusterSingle || clusterType == clusterv1alpha1.ClusterTypeHub {
		/*err := validateHubTypes(poolMap, clusterType)
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
	if true || clusterType == clusterv1alpha1.ClusterTypeClusterSingle || clusterType == clusterv1alpha1.ClusterTypeClusterMulti {
		/*err := validateClusterTypes(poolMap, clusterType)
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
	/*for poolType := range poolMap {
		if _, ok := allowedTypes[poolType]; ok {
			continue
		}
		result = multierror.Append(result, fmt.Errorf("the pool type %s is not supported for a %s", poolType, clusterType))
	}
	return result
}
*/

// validate network configuration
func (c *Cluster) validateNetwork() (result error) {
	// make the choice between deploying into existing VPC or creating a new one
	if _, ok := c.Config().Network.ObjectMeta.Annotations[clusterv1alpha1.ExistingVPCAnnotationKey]; ok {
		// TODO: handle existing vpc
		_, net, err := net.ParseCIDR(c.Config().Network.CIDR)
		if err != nil {
			return fmt.Errorf("error parsing network: %s", err)
		}
		c.networkCIDR = net
	} else {
		_, net, err := net.ParseCIDR(c.Config().Network.CIDR)
		if err != nil {
			return fmt.Errorf("error parsing network: %s", err)
		}
		c.networkCIDR = net
	}

	return nil
}

// validate logging configuration
func (c *Cluster) validateLoggingSinks() (result error) {

	if c.Config().LoggingSinks != nil {
		for index, loggingSink := range c.Config().LoggingSinks {
			if loggingSink.ElasticSearch != nil && loggingSink.ElasticSearch.AmazonESProxy != nil {
				if loggingSink.ElasticSearch.HTTPBasicAuth != nil {
					return fmt.Errorf("cannot enable AWS elasticsearch proxy and HTTP basic auth for logging sink %d", index)
				}
				if loggingSink.ElasticSearch.TLSVerify {
					return fmt.Errorf("cannot enable AWS elasticsearch proxy and force certificate validation for logging sink %d", index)
				}
				if loggingSink.ElasticSearch.TLSCA != "" {
					return fmt.Errorf("cannot enable AWS elasticsearch proxy and specify a custom CA for logging sink %d", index)
				}
			}
		}
	}

	return nil
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

func (c *Cluster) RemoteState() string {
	return c.Environment().Provider().RemoteState(c.Environment().Name(), c.Name(), "main")
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
	if c.Type() == clusterv1alpha1.ClusterTypeClusterMulti {
		return filepath.Join(c.Environment().Tarmak().ConfigPath(), c.Environment().HubName(), "ssh_config")
	}
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

	imageIDs, err := c.ImageIDs()
	if err != nil {
		c.log.Fatalf("error getting image IDs: %s", err)
	}

	// publish instance count and ami ids per instance pool
	for _, instancePool := range c.InstancePools() {
		image := instancePool.Image()
		ids, ok := imageIDs[image]
		if ok {
			output[fmt.Sprintf("%s_ami", instancePool.TFName())] = ids
		}
		output[fmt.Sprintf("%s_instance_count", instancePool.TFName())] = instancePool.Config().MinCount
	}

	// set network cidr
	if c.networkCIDR != nil {
		output["network"] = c.networkCIDR
	}

	key, ok := c.Config().Network.ObjectMeta.Annotations[clusterv1alpha1.ExistingVPCAnnotationKey]
	if ok {
		output["vpc_id"] = key
	}

	privateSubnetIDs, ok := c.Config().Network.ObjectMeta.Annotations[clusterv1alpha1.ExistingPrivateSubnetIDsAnnotationKey]
	if ok {
		output["private_subnets"] = privateSubnetIDs
	}

	publicSubnetIDs, ok := c.Config().Network.ObjectMeta.Annotations[clusterv1alpha1.ExistingPublicSubnetIDsAnnotationKey]
	if ok {
		output["public_subnets"] = publicSubnetIDs
	}

	// publish changed private zone
	if privateZone := c.Environment().Config().PrivateZone; privateZone != "" {
		output["private_zone"] = privateZone
	}

	output["name"] = c.Name()

	return output

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
}

// SetState records the state of Terraform
func (c *Cluster) SetState(state string) {
	c.state = state
}

// GetState retreives the state of Terraform
func (c *Cluster) GetState() string {
	return c.state
}

// get the terrform output from this cluster
func (c *Cluster) TerraformOutput() (map[string]interface{}, error) {
	return c.Environment().Tarmak().Terraform().Output(c)
}

// get API server public hostname
func (c *Cluster) PublicAPIHostname() string {
	if c.conf.Kubernetes == nil || c.conf.Kubernetes.APIServer == nil || c.conf.Kubernetes.APIServer.Public == false {
		return ""
	}

	return fmt.Sprintf(
		"api.%s-%s.%s",
		c.Environment().Name(),
		c.Name(),
		c.Environment().Provider().PublicZone(),
	)
}
