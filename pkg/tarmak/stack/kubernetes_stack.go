package stack

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jetstack-experimental/vault-helper/pkg/kubernetes"

	"github.com/jetstack/tarmak/pkg/tarmak/config"
	"github.com/jetstack/tarmak/pkg/tarmak/interfaces"
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

	s.name = config.StackNameKubernetes
	s.verifyPre = append(s.verifyPre, k.ensureVaultSetup)
	s.verifyPre = append(s.verifyPre, k.ensurePuppetTarGz)

	return k, nil
}

func (s *KubernetesStack) Variables() map[string]interface{} {
	if s.initTokens != nil {
		return s.initTokens
	}
	return map[string]interface{}{}
}

func (s *KubernetesStack) ensurePuppetTarGz() error {
	t := s.Context().Environment().Tarmak()

	rootPath, err := t.RootPath()
	if err != nil {
		return fmt.Errorf("error getting rootPath: %s", err)
	}

	path := filepath.Join(rootPath, "terraform", "aws-centos", "kubernetes", "puppet.tar.gz")

	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		return fmt.Errorf("error creating %s: %s", path, err)
	}

	if err = t.Puppet().TarGz(file); err != nil {
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

	vaultClient := vaultTunnel.VaultClient()

	vaultRootToken, err := s.Context().Environment().VaultRootToken()
	if err != nil {
		return err
	}

	vaultClient.SetToken(vaultRootToken)

	k := kubernetes.New(vaultClient)
	k.SetClusterID(s.Context().ContextName())

	if err := vaultTunnel.Start(); err != nil {
		return err
	}
	defer vaultTunnel.Stop()

	if err := k.Ensure(); err != nil {
		return err
	}

	s.initTokens = map[string]interface{}{}
	for role, token := range k.InitTokens() {
		s.initTokens[fmt.Sprintf("vault_init_token_%s", role)] = token
	}

	return nil
}
