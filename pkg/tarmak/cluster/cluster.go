// Copyright Jetstack Ltd. See LICENSE for details.
package cluster

import (
	"errors"
	"fmt"
	"net"
	"path/filepath"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/go-version"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/util/sets"

	clusterv1alpha1 "github.com/jetstack/tarmak/pkg/apis/cluster/v1alpha1"
	"github.com/jetstack/tarmak/pkg/tarmak/instance_pool"
	"github.com/jetstack/tarmak/pkg/tarmak/interfaces"
	"github.com/jetstack/tarmak/pkg/tarmak/role"
	wingclient "github.com/jetstack/tarmak/pkg/wing/client/clientset/versioned"
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
	ctx  interfaces.CancellationContext

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
		ctx:         environment.Tarmak().CancellationContext(),
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
		if apiServer := k.APIServer; apiServer != nil {
			if master := cluster.Role("master"); master != nil {

				if apiServer.Public == true {
					master.AWS.ELBAPIPublic = true
					if a := apiServer.Amazon; a != nil && a.PublicELBAccessLogs != nil {
						master.AWS.EnablePublicELBAccessLogs = *a.PublicELBAccessLogs.Enabled
					}
				}

				if a := apiServer.Amazon; a != nil && a.InternalELBAccessLogs != nil {
					master.AWS.EnableInternalELBAccessLogs = *a.InternalELBAccessLogs.Enabled
				}
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

func (c *Cluster) validateClusterInstancePoolTypes() error {
	errMap := make(map[string]bool)
	poolMap := make(map[string]bool)

	switch c.Type() {
	case clusterv1alpha1.ClusterTypeHub:
		poolMap = map[string]bool{
			clusterv1alpha1.InstancePoolTypeVault:   true,
			clusterv1alpha1.InstancePoolTypeBastion: true,
			clusterv1alpha1.InstancePoolTypeJenkins: true,
		}

		break

	case clusterv1alpha1.ClusterTypeClusterMulti:
		poolMap = map[string]bool{
			clusterv1alpha1.InstancePoolTypeMaster:     true,
			clusterv1alpha1.InstancePoolTypeWorker:     true,
			clusterv1alpha1.InstancePoolTypeEtcd:       true,
			clusterv1alpha1.InstancePoolTypeMasterEtcd: true,
			clusterv1alpha1.InstancePoolTypeHybrid:     true,
			clusterv1alpha1.InstancePoolTypeAll:        true,
		}

		break

	case clusterv1alpha1.ClusterTypeClusterSingle:
		poolMap = map[string]bool{
			clusterv1alpha1.InstancePoolTypeMaster:     true,
			clusterv1alpha1.InstancePoolTypeWorker:     true,
			clusterv1alpha1.InstancePoolTypeEtcd:       true,
			clusterv1alpha1.InstancePoolTypeBastion:    true,
			clusterv1alpha1.InstancePoolTypeJenkins:    true,
			clusterv1alpha1.InstancePoolTypeVault:      true,
			clusterv1alpha1.InstancePoolTypeAll:        true,
			clusterv1alpha1.InstancePoolTypeMasterEtcd: true,
			clusterv1alpha1.InstancePoolTypeHybrid:     true,
		}

		break

	default:
		return fmt.Errorf("cluster type '%s' not supported", c.Type())
	}

	for _, i := range c.Config().InstancePools {
		if _, ok := poolMap[i.Type]; !ok {
			errMap[i.Type] = true
		}
	}

	var result *multierror.Error
	for t := range errMap {
		err := fmt.Errorf("instance pool type '%s' not valid in cluster type '%s'", t, c.Type())
		result = multierror.Append(result, err)
	}

	return result.ErrorOrNil()
}

func (c *Cluster) validateSingleInstancePoolMap(poolMap map[string][]*clusterv1alpha1.InstancePool, singleList []string) error {
	var result *multierror.Error

	for _, i := range singleList {
		if v, ok := poolMap[i]; !ok || len(v) != 1 {
			err := fmt.Errorf("cluster type '%s' requires exactly one '%s' instance pool", c.Type(), i)
			result = multierror.Append(result, err)
		}
	}

	return result.ErrorOrNil()
}

func (c *Cluster) validateMultiInstancePoolMap(poolMap map[string][]*clusterv1alpha1.InstancePool, instanceType string) error {
	if len(poolMap[instanceType]) < 1 {
		return fmt.Errorf("cluster type '%s' requires one or more instance pool of type '%s'", c.Type(), instanceType)
	}

	return nil
}

func (c *Cluster) validateClusterInstancePoolCount() error {
	var result *multierror.Error

	poolMap := make(map[string][]*clusterv1alpha1.InstancePool)
	for _, i := range c.Config().InstancePools {
		poolMap[i.Type] = append(poolMap[i.Type], &i)
	}

	if c.Type() != clusterv1alpha1.ClusterTypeHub {
		err := c.validateMultiInstancePoolMap(poolMap, clusterv1alpha1.InstancePoolTypeWorker)
		if err != nil {
			result = multierror.Append(result, err)
		}
	}

	switch c.Type() {
	case clusterv1alpha1.ClusterTypeClusterSingle:
		if err := c.validateSingleInstancePoolMap(poolMap, []string{
			clusterv1alpha1.InstancePoolTypeBastion,
			clusterv1alpha1.InstancePoolTypeVault,
			clusterv1alpha1.InstancePoolTypeEtcd,
		}); err != nil {
			result = multierror.Append(result, err)
		}

		if err := c.validateMultiInstancePoolMap(poolMap, clusterv1alpha1.InstancePoolTypeMaster); err != nil {
			result = multierror.Append(result, err)
		}

		break

	case clusterv1alpha1.ClusterTypeHub:

		if err := c.validateSingleInstancePoolMap(poolMap, []string{
			clusterv1alpha1.InstancePoolTypeBastion,
			clusterv1alpha1.InstancePoolTypeVault,
		}); err != nil {
			result = multierror.Append(result, err)
		}

		break

	case clusterv1alpha1.ClusterTypeClusterMulti:
		if err := c.validateSingleInstancePoolMap(poolMap, []string{
			clusterv1alpha1.InstancePoolTypeEtcd,
		}); err != nil {
			result = multierror.Append(result, err)
		}

		if err := c.validateMultiInstancePoolMap(poolMap, clusterv1alpha1.InstancePoolTypeMaster); err != nil {
			result = multierror.Append(result, err)
		}

		break

	default:
		return fmt.Errorf("cluster type '%s' is not a supported type", c.Type())
	}

	return result.ErrorOrNil()
}

// validate server pools
func (c *Cluster) validateInstancePools() error {
	var result *multierror.Error

	for _, instancePool := range c.InstancePools() {
		err := instancePool.Validate()
		if err != nil {
			result = multierror.Append(result, err)
		}
	}

	if result.ErrorOrNil() != nil {
		return result.ErrorOrNil()
	}

	// validate instance pool types according to cluster type
	if err := c.validateClusterInstancePoolTypes(); err != nil {
		return err
	}

	// validate instance pool count according to cluster type
	if err := c.validateClusterInstancePoolCount(); err != nil {
		return err
	}

	if err := c.validateSubnets(); err != nil {
		return err
	}

	return nil
}

// Verify cluster
func (c *Cluster) Verify() error {
	var result *multierror.Error

	if err := c.Environment().Verify(); err != nil {
		return fmt.Errorf("failed to verify tarmak provider: %s", err)
	}

	if err := c.VerifyInstancePools(); err != nil {
		result = multierror.Append(result, err)
	}

	if c.Type() == clusterv1alpha1.ClusterTypeClusterMulti {
		if err := c.verifyHubState(); err != nil {
			result = multierror.Append(result, err)
		}
	}

	return result.ErrorOrNil()
}

func (c *Cluster) verifyHubState() error {
	// The hub should be manually applied first to ensure the vault token and private key can be saved
	errMsg := "hub cluster must be applied once first"
	err := c.Environment().Tarmak().Terraform().Prepare(c.Environment().Hub())
	if err != nil {
		return fmt.Errorf("failed to prepare hub cluster for output, %s: %v", errMsg, err)
	}
	output, err := c.Environment().Tarmak().Terraform().Output(c.Environment().Hub())
	if err != nil {
		return fmt.Errorf("failed to get hub cluster output values, %s: %v", errMsg, err)
	}

	requiredHubResources := []string{
		"bastion_bastion_instance_id",
		"bastion_bastion_security_group_id",
		"instance_fqdns",
		"network_availability_zones",
		"network_private_subnet_ids",
		"network_private_zone",
		"network_private_zone_id",
		"network_public_subnet_ids",
		"network_vpc_id",
		"state_public_zone",
		"state_public_zone_id",
		"state_secrets_bucket",
		"vault_ca",
		"vault_instance_fqdns",
		"vault_vault_ca",
		"vault_vault_kms_key_id",
		"vault_vault_security_group_id",
		"vault_vault_unseal_key_name",
		"vault_vault_url",
	}
	var result *multierror.Error
	for _, r := range requiredHubResources {
		o, ok := output[r]
		if !ok || o == nil {
			err := fmt.Errorf("'%s' not found", r)
			result = multierror.Append(result, err)
		}
	}

	if result.ErrorOrNil() != nil {
		return fmt.Errorf("required hub cluster resource(s) not found, %s: %v", errMsg, result.ErrorOrNil())
	}

	return nil
}

// Verify instance pools
func (c *Cluster) VerifyInstancePools() error {
	imageIDs, err := c.ImageIDs()
	if err != nil {
		return fmt.Errorf("error getting image IDs: %s]", err)
	}

	if len(imageIDs) == 0 {
		return errors.New("no images found, please run `$ tarmak cluster images build`")
	}

	var result *multierror.Error
	for _, instancePool := range c.InstancePools() {
		image := instancePool.Image()
		s, ok := imageIDs[image]
		if !ok || s == "" {
			err := fmt.Errorf("failed to find the image ID of image '%s' used by instance pool '%s'", image, instancePool.TFName())
			result = multierror.Append(result, err)
		}
	}

	return result.ErrorOrNil()
}

func (c *Cluster) Validate() error {
	var result *multierror.Error

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

		//validate prometheus mode
		if c.Config().Kubernetes.Prometheus != nil {
			if err := c.validatePrometheusMode(); err != nil {
				result = multierror.Append(result, err)
			}
		}
	}

	return result.ErrorOrNil()
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
			if loggingSink.Elasticsearch != nil && loggingSink.Elasticsearch.AmazonESProxy != nil {
				if loggingSink.Elasticsearch.HTTPBasicAuth != nil {
					return fmt.Errorf("cannot enable AWS elasticsearch proxy and HTTP basic auth for logging sink %d", index)
				}
				if loggingSink.Elasticsearch.TLSVerify {
					return fmt.Errorf("cannot enable AWS elasticsearch proxy and force certificate validation for logging sink %d", index)
				}
				if loggingSink.Elasticsearch.TLSCA != "" {
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
		if !c.Config().Kubernetes.ClusterAutoscaler.Overprovisioning.Enabled {
			return nil
		}
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

		if s := c.Config().Kubernetes.ClusterAutoscaler.ScaleDownUtilizationThreshold; s != nil && (*s < 0 || *s > 1) {
			return fmt.Errorf("scale down threshold '%v' unacceptable, must be value between 0 and 1", *c.Config().Kubernetes.ClusterAutoscaler.ScaleDownUtilizationThreshold)
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

	api := c.Config().Kubernetes.APIServer
	if a := api.Amazon; a != nil {
		for _, l := range []*clusterv1alpha1.ClusterKubernetesAPIServerAmazonAccessLogs{
			a.PublicELBAccessLogs,
			a.InternalELBAccessLogs,
		} {

			if l != nil && *l.Enabled {
				if len(l.Bucket) == 0 {
					result = multierror.Append(result, errors.New("access logs enabled with no bucket name"))
				}

				if *l.Interval != 5 && *l.Interval != 60 {
					result = multierror.Append(result, errors.New("access logs interval may only be a value of 5 or 60"))
				}
			}

		}
	}

	vK, err := version.NewVersion(c.Config().Kubernetes.Version)
	if err != nil {
		return multierror.Append(result, err)
	}

	v11, err := version.NewVersion("1.11.0")
	if err != nil {
		return multierror.Append(result, err)
	}

	if vK.LessThan(v11) && len(api.DisableAdmissionControllers) > 0 {
		return multierror.Append(result, fmt.Errorf(
			"kubernetes version less than 1.11.0 expects no disable admission controllers, found: %s",
			api.DisableAdmissionControllers))
	}

	return result
}

func (c *Cluster) validatePrometheusMode() error {
	var result error

	allowedModes := sets.NewString(
		clusterv1alpha1.PrometheusModeFull,
		clusterv1alpha1.PrometheusModeExternalScrapeTargetsOnly,
		clusterv1alpha1.PrometheusModeExternalExportersOnly,
	)

	modeString := c.Config().Kubernetes.Prometheus.Mode
	if c.Config().Kubernetes.Prometheus != nil && modeString != "" {
		if !allowedModes.Has(modeString) {
			return fmt.Errorf("%s is not a valid Prometheus mode, allowed modes: %s", modeString, allowedModes.List())
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
		imageMap, err := c.Environment().Tarmak().Packer().IDs(c.AmazonEBSEncrypted())
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
		output[fmt.Sprintf("%s_root_volume_size", instancePool.TFName())] = instancePool.RootVolume().Size()
		output[fmt.Sprintf("%s_root_volume_type", instancePool.TFName())] = instancePool.RootVolume().Type()
		output[fmt.Sprintf("%s_iam_additional_policy_arns", instancePool.TFName())] = instancePool.Config().Amazon.AdditionalIAMPolicies
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

	// Get enabled elb access logs
	if k := c.Config().Kubernetes; k != nil && k.APIServer != nil && k.APIServer.Amazon != nil {
		if p := k.APIServer.Amazon.PublicELBAccessLogs; p != nil {
			output["elb_access_logs_public_enabled"] = fmt.Sprintf("%v", *p.Enabled)
			output["elb_access_logs_public_bucket"] = p.Bucket
			output["elb_access_logs_public_bucket_prefix"] = p.BucketPrefix
			output["elb_access_logs_public_bucket_interval"] = *p.Interval
		} else {
			output["elb_access_logs_public_enabled"] = "false"
		}

		if i := k.APIServer.Amazon.InternalELBAccessLogs; i != nil {
			output["elb_access_logs_internal_enabled"] = fmt.Sprintf("%v", *i.Enabled)
			output["elb_access_logs_internal_bucket"] = i.Bucket
			output["elb_access_logs_internal_bucket_prefix"] = i.BucketPrefix
			output["elb_access_logs_internal_bucket_interval"] = *i.Interval
		} else {
			output["elb_access_logs_internal_enabled"] = "false"
		}
	} else {
		output["elb_access_logs_public_enabled"] = "false"
		output["elb_access_logs_internal_enabled"] = "false"
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

// retrieve Amazons EBS encryption status
func (c *Cluster) AmazonEBSEncrypted() bool {
	if a := c.conf.Amazon; a != nil && a.EBSEncrypted != nil {
		return *a.EBSEncrypted
	}
	return false
}

func (c *Cluster) validateSubnets() error {
	var result *multierror.Error

	if c.Type() == clusterv1alpha1.ClusterTypeClusterMulti && c.Environment().Hub() != nil {
		hSubnets := c.Environment().Hub().Subnets()

		for _, cNet := range c.Subnets() {
			found := false

			for _, hNet := range hSubnets {
				if cNet.Zone == hNet.Zone {
					found = true
					break
				}
			}

			if !found {
				err := fmt.Errorf("hub cluster does not include zone '%s'", cNet.Zone)
				result = multierror.Append(result, err)
			}
		}
	}

	return result.ErrorOrNil()
}
