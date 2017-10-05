package stack

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jetstack-experimental/vault-helper/pkg/kubernetes"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	clusterv1alpha1 "github.com/jetstack/tarmak/pkg/apis/cluster/v1alpha1"
	tarmakv1alpha1 "github.com/jetstack/tarmak/pkg/apis/tarmak/v1alpha1"
	"github.com/jetstack/tarmak/pkg/tarmak/interfaces"
)

type KubernetesStack struct {
	*Stack
	initTokens map[string]interface{}
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
	s.verifyPostDeploy = append(s.verifyPreDeploy, k.ensurePuppetConverged)
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

func (s *KubernetesStack) ensurePuppetConverged() error {
	// list all instances in Provider
	providerInstances, err := s.Cluster().ListHosts()
	providerInstaceMap := make(map[string]bool)
	for _, instance := range providerInstances {
		providerInstaceMap[instance.ID()] = true
	}

	// connect to wing
	clientset, tunnel, err := s.Cluster().Environment().WingClientset()
	if err != nil {
		return fmt.Errorf("failed to connect to wing API on bastion: %s", err)
	}
	defer tunnel.Stop()
	client := clientset.WingV1alpha1().Instances(s.Cluster().ClusterName())

	// list all instances in wing
	wingInstances, err := client.List(metav1.ListOptions{})
	if err != nil {
		return err
	}

	for _, instance := range wingInstances.Items {
		if _, ok := providerInstaceMap[instance.Name]; !ok {
			s.log.Debugf("deleting unused instance %s in wing API", instance.Name)
			if err := client.Delete(instance.Name, &metav1.DeleteOptions{}); err != nil {
				s.log.Warnf("error deleting instance %s in wing API: %s", instance.Name, err)
			}
		}
	}

	return nil
}
