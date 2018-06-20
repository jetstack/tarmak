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

const (
	// represents Terraform in a destroy state
	StateDestroy                          = "destroy"
	ExistingVPCAnnotationKey              = "tarmak.io/existing-vpc-id"
	ExistingPublicSubnetIDsAnnotationKey  = "tarmak.io/existing-public-subnet-ids"
	ExistingPrivateSubnetIDsAnnotationKey = "tarmak.io/existing-private-subnet-ids"
	JenkinsCertificateARNAnnotationKey    = "tarmak.io/jenkins-certificate-arn"
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

	if err := cluster.Validate(); err != nil {
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
	return c.Environment().Provider().ListHosts(c)
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
	for _, instancePool := range c.InstancePools() {
		err := instancePool.Validate()
		if err != nil {
			result = multierror.Append(result, err)
		}
	}
	return result
}

// Verify cluster
func (c *Cluster) Verify() (result error) {
	return c.VerifyInstancePools()
}

// Verify instance pools
func (c *Cluster) VerifyInstancePools() (result error) {
	imageIDs, err := c.ImageIDs()
	if err != nil {
		return fmt.Errorf("error getting image IDs: %s]", err)
	}

	for _, instancePool := range c.InstancePools() {
		image := instancePool.Image()
		_, ok := imageIDs[image]
		if !ok {
			return fmt.Errorf("error getting the image ID of %s", instancePool.TFName())
		}
	}
	return nil
}

func (c *Cluster) Validate() (result error) {
	// validate instance pools
	if err := c.validateInstancePools(); err != nil {
		result = multierror.Append(result, err)
	}

	// validate network setup
	if err := c.validateNetwork(); err != nil {
		result = multierror.Append(result, err)
	}

	//validate logging
	if err := c.validateLoggingSinks(); err != nil {
		result = multierror.Append(result, err)
	}

	// validate overprovisioning
	if err := c.validateClusterAutoscaler(); err != nil {
		result = multierror.Append(result, fmt.Errorf("invalid overprovisioning configuration: %s", err))
	}

	//validate apiserver
	if k := c.Config().Kubernetes; k != nil {
		if apiServer := k.APIServer; apiServer != nil {
			if err := c.validateAPIServer(); err != nil {
				result = multierror.Append(result, err)
			}
		}
	}

	return result
}

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

// validate overprovisioning
func (c *Cluster) validateClusterAutoscaler() (result error) {

	if c.Config().Kubernetes != nil && c.Config().Kubernetes.ClusterAutoscaler != nil && c.Config().Kubernetes.ClusterAutoscaler.Overprovisioning != nil {
		if c.Config().Kubernetes.ClusterAutoscaler.Overprovisioning.Enabled && !c.Config().Kubernetes.ClusterAutoscaler.Enabled {
			return fmt.Errorf("cannot enable overprovisioning if cluster autoscaling is disabled")
		}
		if c.Config().Kubernetes.ClusterAutoscaler.Overprovisioning.ReservedMegabytesPerReplica < 0 ||
			c.Config().Kubernetes.ClusterAutoscaler.Overprovisioning.ReservedMillicoresPerReplica < 0 ||
			c.Config().Kubernetes.ClusterAutoscaler.Overprovisioning.CoresPerReplica < 0 ||
			c.Config().Kubernetes.ClusterAutoscaler.Overprovisioning.NodesPerReplica < 0 ||
			c.Config().Kubernetes.ClusterAutoscaler.Overprovisioning.ReplicaCount < 0 {
			return fmt.Errorf("cannot set negative overprovisioning parameters")
		}
		if c.Config().Kubernetes.ClusterAutoscaler.Overprovisioning.ReservedMegabytesPerReplica == 0 && c.Config().Kubernetes.ClusterAutoscaler.Overprovisioning.ReservedMillicoresPerReplica == 0 {
			return fmt.Errorf("one of reservedMillicoresPerReplica and reservedMegabytesPerReplica must be set")
		}
		if (c.Config().Kubernetes.ClusterAutoscaler.Overprovisioning.CoresPerReplica > 0 || c.Config().Kubernetes.ClusterAutoscaler.Overprovisioning.NodesPerReplica > 0) && c.Config().Kubernetes.ClusterAutoscaler.Overprovisioning.ReplicaCount > 0 {
			return fmt.Errorf("cannot configure both static and per replica overprovisioning rules")
		}
		if (c.Config().Kubernetes.ClusterAutoscaler.Overprovisioning.Image != "" || c.Config().Kubernetes.ClusterAutoscaler.Overprovisioning.Version != "") && (c.Config().Kubernetes.ClusterAutoscaler.Overprovisioning.CoresPerReplica == 0 && c.Config().Kubernetes.ClusterAutoscaler.Overprovisioning.NodesPerReplica == 0) {
			return fmt.Errorf("setting overprovisioning image or version is only valid when proportional overprovisioning is enabled")
		}
	}

	return nil
}

// Validate APIServer
func (c *Cluster) validateAPIServer() (result error) {
	for _, cidr := range c.Config().Kubernetes.APIServer.AllowCIDRs {
		_, _, err := net.ParseCIDR(cidr)
		if err != nil {
			result = multierror.Append(result, fmt.Errorf("%s is not a valid CIDR format", cidr))
		}
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
		if instancePool.Config().AllowCIDRs != nil {
			output[fmt.Sprintf("%s_admin_cidrs", instancePool.TFName())] = instancePool.Config().AllowCIDRs
		} else {
			output[fmt.Sprintf("%s_admin_cidrs", instancePool.TFName())] = c.environment.Config().AdminCIDRs
		}
		output[fmt.Sprintf("%s_min_instance_count", instancePool.TFName())] = instancePool.Config().MinCount
		output[fmt.Sprintf("%s_max_instance_count", instancePool.TFName())] = instancePool.Config().MaxCount
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

	for _, instancePool := range c.InstancePools() {
		if instancePool.Role().Name() == "jenkins" {
			jenkinsCertificateARN, ok := instancePool.Config().Annotations[JenkinsCertificateARNAnnotationKey]
			if ok {
				output["jenkins_certificate_arn"] = jenkinsCertificateARN
				break
			}
		}
	}

	// Get Apiserver valid admin cidrs
	if k := c.Config().Kubernetes; k != nil {
		if apiServer := k.APIServer; apiServer != nil && apiServer.AllowCIDRs != nil {
			output["api_admin_cidrs"] = apiServer.AllowCIDRs
		} else {
			output["api_admin_cidrs"] = c.environment.Config().AdminCIDRs
		}
	} else {
		output["api_admin_cidrs"] = c.environment.Config().AdminCIDRs
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
