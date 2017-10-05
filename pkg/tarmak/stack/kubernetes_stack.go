package stack

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jetstack-experimental/vault-helper/pkg/kubernetes"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	clusterv1alpha1 "github.com/jetstack/tarmak/pkg/apis/cluster/v1alpha1"
	tarmakv1alpha1 "github.com/jetstack/tarmak/pkg/apis/tarmak/v1alpha1"
	wingv1alpha1 "github.com/jetstack/tarmak/pkg/apis/wing/v1alpha1"
	"github.com/jetstack/tarmak/pkg/tarmak/interfaces"
	wingclient "github.com/jetstack/tarmak/pkg/wing/client"
	wingclientv1alpha1 "github.com/jetstack/tarmak/pkg/wing/client/typed/wing/v1alpha1"
)

type KubernetesStack struct {
	*Stack
	initTokens map[string]interface{}

	wingClientset *wingclient.Clientset
	wingTunnel    interfaces.Tunnel
}

var _ interfaces.Stack = &KubernetesStack{}

func newKubernetesStack(s *Stack) (*KubernetesStack, error) {
	k := &KubernetesStack{
		Stack: s,
	}

	s.roles = make(map[string]bool)
	s.roles[clusterv1alpha1.InstancePoolTypeEtcd] = true
	s.roles[clusterv1alpha1.InstancePoolTypeMaster] = true
	s.roles[clusterv1alpha1.InstancePoolTypeWorker] = true

	s.name = tarmakv1alpha1.StackNameKubernetes
	s.verifyPreDeploy = append(s.verifyPreDeploy, k.ensureVaultSetup)
	s.verifyPreDeploy = append(s.verifyPreDeploy, k.ensurePuppetTarGz)
	s.verifyPostDeploy = append(s.verifyPostDeploy, k.ensurePuppetReapply)
	s.verifyPostDeploy = append(s.verifyPostDeploy, k.ensurePuppetConverged)
	s.verifyPreDestroy = append(s.verifyPreDestroy, k.emptyPuppetTarGz)

	return k, nil
}

func (s *KubernetesStack) Variables() map[string]interface{} {
	vars := s.Stack.Variables()

	if s.initTokens != nil {
		for key, val := range s.initTokens {
			vars[key] = val
		}
	}
	return vars
}

func (s *KubernetesStack) puppetTarGzPath() (string, error) {
	rootPath, err := s.Cluster().Environment().Tarmak().RootPath()
	if err != nil {
		return "", fmt.Errorf("error getting rootPath: %s", err)
	}

	path := filepath.Join(rootPath, "terraform", s.Cluster().Environment().Provider().Cloud(), "kubernetes", "puppet.tar.gz")

	return path, nil
}

func (s *KubernetesStack) emptyPuppetTarGz() error {
	path, err := s.puppetTarGzPath()
	if err != nil {
		return err
	}

	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		return fmt.Errorf("error creating %s: %s", path, err)
	}

	if err := file.Close(); err != nil {
		return fmt.Errorf("error closing %s: %s", path, err)
	}

	return nil

}

func (s *KubernetesStack) ensurePuppetTarGz() error {
	path, err := s.puppetTarGzPath()
	if err != nil {
		return err
	}

	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		return fmt.Errorf("error creating %s: %s", path, err)
	}

	if err = s.Cluster().Environment().Tarmak().Puppet().TarGz(file); err != nil {
		return fmt.Errorf("error writing to %s: %s", path, err)
	}

	if err := file.Close(); err != nil {
		return fmt.Errorf("error closing %s: %s", path, err)
	}

	return nil

}

func (s *KubernetesStack) ensureVaultSetup() error {
	vaultStack := s.Cluster().Environment().VaultStack()

	// load outputs from terraform
	s.Cluster().Environment().Tarmak().Terraform().Output(vaultStack)

	vaultStackReal, ok := vaultStack.(*VaultStack)
	if !ok {
		return fmt.Errorf("unexpected type for vault stack: %T", vaultStack)
	}

	vaultTunnel, err := vaultStackReal.VaultTunnel()
	if err != nil {
		return err
	}
	defer vaultTunnel.Stop()

	vaultClient := vaultTunnel.VaultClient()

	vaultRootToken, err := s.Cluster().Environment().VaultRootToken()
	if err != nil {
		return err
	}

	vaultClient.SetToken(vaultRootToken)

	k := kubernetes.New(vaultClient)
	k.SetClusterID(s.Cluster().ClusterName())

	if err := k.Ensure(); err != nil {
		return err
	}

	s.initTokens = map[string]interface{}{}
	for role, token := range k.InitTokens() {
		s.initTokens[fmt.Sprintf("vault_init_token_%s", role)] = token
	}
	return nil
}

func (s *KubernetesStack) wingInstanceClient() (wingclientv1alpha1.InstanceInterface, error) {
	var err error

	if s.wingClientset == nil {
		// connect to wing
		s.wingClientset, s.wingTunnel, err = s.Cluster().Environment().WingClientset()
		if err != nil {
			return nil, fmt.Errorf("failed to connect to wing API on bastion: %s", err)
		}
	}

	return s.wingClientset.WingV1alpha1().Instances(s.Cluster().ClusterName()), nil
}

func (s *KubernetesStack) listInstances() (instances []*wingv1alpha1.Instance, err error) {
	// connect to wing
	client, err := s.wingInstanceClient()
	if err != nil {
		return instances, fmt.Errorf("failed to connect to wing API on bastion: %s", err)
	}

	// list all instances in Provider
	providerInstances, err := s.Cluster().ListHosts()
	providerInstaceMap := make(map[string]interfaces.Host)
	if err != nil {
		return instances, fmt.Errorf("failed to list provider's instances: %s", err)
	}

	for pos, _ := range providerInstances {
		providerInstaceMap[providerInstances[pos].ID()] = providerInstances[pos]
	}

	// list all instances in wing
	wingInstances, err := client.List(metav1.ListOptions{})
	if err != nil {
		return instances, err
	}

	// loop through instances
	for pos, _ := range wingInstances.Items {
		instance := &wingInstances.Items[pos]

		// removes instances not in AWS
		if _, ok := providerInstaceMap[instance.Name]; !ok {
			s.log.Debugf("deleting unused instance %s in wing API", instance.Name)
			if err := client.Delete(instance.Name, &metav1.DeleteOptions{}); err != nil {
				s.log.Warnf("error deleting instance %s in wing API: %s", instance.Name, err)
			}
			continue
		}
		instances = append(instances, instance)
	}

	return instances, nil

}

func (s *KubernetesStack) ensurePuppetReapply() error {
	s.log.Debugf("making sure all nodes apply the latest manifest")

	// connect to wing
	client, err := s.wingInstanceClient()
	if err != nil {
		return fmt.Errorf("failed to connect to wing API on bastion: %s", err)
	}

	// list instances
	instances, err := s.listInstances()
	if err != nil {
		return fmt.Errorf("failed to list instances: %s", err)
	}

	for pos, _ := range instances {
		instance := instances[pos]
		if instance.Spec == nil {
			instance.Spec = &wingv1alpha1.InstanceSpec{}
		}
		instance.Spec.Converge = &wingv1alpha1.InstanceSpecManifest{}

		if _, err := client.Update(instance); err != nil {
			s.log.Warnf("error updating instance %s in wing API: %s", instance.Name, err)
		}
	}

	// TODO: solve this on the API server side
	time.Sleep(time.Second * 5)

	return nil
}

func (s *KubernetesStack) ensurePuppetConverged() error {
	s.log.Debugf("making sure all nodes have converged using puppet")

	retries := 50
	for {
		instances, err := s.listInstances()
		if err != nil {
			return fmt.Errorf("failed to list instances: %s", err)
		}

		instanceByState := make(map[wingv1alpha1.InstanceManifestState][]*wingv1alpha1.Instance)

		for pos, _ := range instances {
			instance := instances[pos]

			// index by instance convergance state
			if instance.Status == nil || instance.Status.Converge == nil || instance.Status.Converge.State == "" {
				continue
			}

			state := instance.Status.Converge.State
			if _, ok := instanceByState[state]; !ok {
				instanceByState[state] = []*wingv1alpha1.Instance{}
			}

			instanceByState[state] = append(
				instanceByState[state],
				instance,
			)
		}

		err = s.checkAllInstancesConverged(instanceByState)
		if err == nil {
			s.log.Info("all instances converged")
			return nil
		} else {
			s.log.Debug(err)
		}

		retries -= 1
		if retries == 0 {
			break
		}
		time.Sleep(time.Second * 5)

	}

	return fmt.Errorf("instances failed to converge in time")
}

func (s *KubernetesStack) checkAllInstancesConverged(byState map[wingv1alpha1.InstanceManifestState][]*wingv1alpha1.Instance) error {
	instancesNotConverged := []*wingv1alpha1.Instance{}
	for key, instances := range byState {
		if len(instances) == 0 {
			continue
		}
		if key != wingv1alpha1.InstanceManifestStateConverged {
			instancesNotConverged = append(instancesNotConverged, instances...)
		}
		s.Log().Debugf("%d instances in state %s: %s", len(instances), key, outputInstances(instances))
	}

	if len(instancesNotConverged) > 0 {
		return fmt.Errorf("not all instances have converged yet %s", outputInstances(instancesNotConverged))
	}

	return nil
}

func outputInstances(instances []*wingv1alpha1.Instance) string {
	var output []string
	for _, instance := range instances {
		output = append(output, instance.Name)
	}
	return strings.Join(output, ", ")
}
