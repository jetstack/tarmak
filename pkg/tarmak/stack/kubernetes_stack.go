package stack

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jetstack-experimental/vault-helper/pkg/kubernetes"

	"github.com/jetstack/tarmak/pkg/tarmak/config"
	"github.com/jetstack/tarmak/pkg/tarmak/interfaces"
	"github.com/jetstack/tarmak/pkg/tarmak/role"
)

type KubernetesStack struct {
	*Stack
	initTokens map[string]interface{}
}

var _ interfaces.Stack = &KubernetesStack{}

func newKubernetesStack(s *Stack, conf *config.StackKubernetes) (*KubernetesStack, error) {
	k := &KubernetesStack{
		Stack: s,
	}

	masterRole := &role.Role{
		Stateful: false,
		AWS: &role.RoleAWS{
			ELBAPI:     true,
			IAMEC2Full: true,
			IAMELBFull: true,
		},
	}
	masterRole.WithName("master").WithPrefix("kubernetes")

	workerRole := &role.Role{
		Stateful: false,
		AWS: &role.RoleAWS{
			ELBIngress:                     true,
			IAMEC2ModifyInstanceAttributes: true,
		},
	}
	workerRole.WithName("worker").WithPrefix("kubernetes")

	etcdRole := &role.Role{
		Stateful: true,
		AWS:      &role.RoleAWS{},
	}
	etcdRole.WithName("etcd").WithPrefix("kubernetes")

	masterEtcdRole := &role.Role{
		Stateful: false,
		AWS: &role.RoleAWS{
			ELBAPI:     true,
			IAMEC2Full: true,
			IAMELBFull: true,
		},
	}
	masterEtcdRole.WithName("master-etcd").WithPrefix("kubernetes")

	s.roles = map[string]*role.Role{
		"master":      masterRole,
		"worker":      workerRole,
		"etcd":        etcdRole,
		"etcd-master": masterEtcdRole,
		"master-etcd": masterEtcdRole,
	}

	s.roles = map[string]*role.Role{
		"master": masterRole,
	}
	s.name = config.StackNameKubernetes
	s.verifyPreDeploy = append(s.verifyPreDeploy, k.ensureVaultSetup)
	s.verifyPreDeploy = append(s.verifyPreDeploy, k.ensurePuppetTarGz)
	s.verifyPreDestroy = append(s.verifyPreDestroy, k.emptyPuppetTarGz)

	return k, nil
}

func (s *KubernetesStack) Variables() map[string]interface{} {
	if s.initTokens != nil {
		return s.initTokens
	}
	return map[string]interface{}{}
}

func (s *KubernetesStack) puppetTarGzPath() (string, error) {
	rootPath, err := s.Context().Environment().Tarmak().RootPath()
	if err != nil {
		return "", fmt.Errorf("error getting rootPath: %s", err)
	}

	path := filepath.Join(rootPath, "terraform", "aws-centos", "kubernetes", "puppet.tar.gz")

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

	if err = s.Context().Environment().Tarmak().Puppet().TarGz(file); err != nil {
		return fmt.Errorf("error writing to %s: %s", path, err)
	}

	if err := file.Close(); err != nil {
		return fmt.Errorf("error closing %s: %s", path, err)
	}

	return nil

}

func (s *KubernetesStack) ensureVaultSetup() error {
	vaultStack := s.Context().Environment().VaultStack()

	// load outputs from terraform
	s.Context().Environment().Tarmak().Terraform().Output(vaultStack)

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

	vaultRootToken, err := s.Context().Environment().VaultRootToken()
	if err != nil {
		return err
	}

	vaultClient.SetToken(vaultRootToken)

	k := kubernetes.New(vaultClient)
	k.SetClusterID(s.Context().ContextName())

	if err := k.Ensure(); err != nil {
		return err
	}

	s.initTokens = map[string]interface{}{}
	for role, token := range k.InitTokens() {
		s.initTokens[fmt.Sprintf("vault_init_token_%s", role)] = token
	}

	return nil
}
