// Copyright Jetstack Ltd. See LICENSE for details.
package kubernetes

import (
	"errors"
	"fmt"
	"path/filepath"
	"time"

	vault "github.com/hashicorp/vault/api"
	"github.com/sirupsen/logrus"
)

type PKIVaultBackend struct {
	pkiName    string
	kubernetes *Kubernetes

	MaxLeaseTTL     time.Duration
	DefaultLeaseTTL time.Duration

	Log *logrus.Entry
}

func NewPKIVaultBackend(k *Kubernetes, pkiName string, logger *logrus.Entry) *PKIVaultBackend {
	return &PKIVaultBackend{
		pkiName:         pkiName,
		kubernetes:      k,
		MaxLeaseTTL:     k.MaxValidityCA,
		DefaultLeaseTTL: k.MaxValidityCA,
		Log:             logger,
	}
}

func (p *PKIVaultBackend) TuneMount(mount *vault.MountOutput) error {
	if p.TuneMountRequired(mount) {
		mountConfig := p.getMountConfigInput()
		err := p.kubernetes.vaultClient.Sys().TuneMount(p.Path(), mountConfig)
		if err != nil {
			return fmt.Errorf("error tuning mount config: %v", err.Error())
		}
		p.Log.Debugf("Tuned Mount: %s", p.pkiName)
		return nil
	}
	p.Log.Debugf("No tune required: %s", p.pkiName)

	return nil
}

func (p *PKIVaultBackend) unMount() error {
	return p.kubernetes.vaultClient.Sys().Unmount(p.Path())
}

func (p *PKIVaultBackend) TuneMountRequired(mount *vault.MountOutput) bool {

	if mount.Config.DefaultLeaseTTL != int(p.DefaultLeaseTTL.Seconds()) {
		return true
	}
	if mount.Config.MaxLeaseTTL != int(p.MaxLeaseTTL.Seconds()) {
		return true
	}

	return false
}

func (p *PKIVaultBackend) Ensure() error {
	mount, err := GetMountByPath(p.kubernetes.vaultClient, p.Path())
	if err != nil {
		return err
	}

	// Mount doesn't Exist
	if mount == nil {
		p.Log.Debugf("No mounts found for: %s", p.pkiName)
		err := p.kubernetes.vaultClient.Sys().Mount(
			p.Path(),
			&vault.MountInput{
				Description: "Kubernetes " + p.kubernetes.clusterID + "/" + p.pkiName + " CA",
				Type:        p.Type(),
			},
		)

		if err != nil {
			return fmt.Errorf("failed to create mount: %v", err)
		}
		mount, err = GetMountByPath(p.kubernetes.vaultClient, p.Path())
		if err != nil {
			return err
		}
		p.Log.Infof("Mounted '%s'", p.pkiName)

	} else {
		if mount.Type != p.Type() {
			return fmt.Errorf("Mount '%s' already existing with wrong type '%s'", p.Path(), mount.Type)
		}
		p.Log.Debugf("Mount '%s' already existing", p.Path())
	}

	if mount != nil {
		err = p.TuneMount(mount)
		if err != nil {
			return errors.New("failed to tune mount")
		}
	}

	return p.ensureCA()
}

func (p *PKIVaultBackend) Delete() error {
	if err := p.unMount(); err != nil {
		return err
	}

	return nil
}

func (p *PKIVaultBackend) EnsureDryRun() (bool, error) {
	mount, err := GetMountByPath(p.kubernetes.vaultClient, p.Path())
	if err != nil {
		return false, err
	}

	// Mount doesn't Exist
	if mount == nil || mount.Type != p.Type() || p.TuneMountRequired(mount) {
		return true, nil
	}

	exist, err := p.caPathExists()
	if err != nil {
		return false, err
	}

	if !exist {
		return true, nil
	}

	return false, nil
}

func (p *PKIVaultBackend) ensureCA() error {
	b, err := p.caPathExists()
	if err != nil {
		return err
	}

	if !b {
		return p.generateCA()
	}

	return nil
}

func (p *PKIVaultBackend) generateCA() error {
	description := "Kubernetes " + p.kubernetes.clusterID + "/" + p.pkiName + " CA"

	data := map[string]interface{}{
		"common_name":          description,
		"ttl":                  p.getMaxLeaseTTL(),
		"exclude_cn_from_sans": true,
	}

	_, err := p.kubernetes.vaultClient.Logical().Write(p.caGenPath(), data)
	if err != nil {
		return fmt.Errorf("error writing new CA: %v", err)
	}

	return nil
}

func (p *PKIVaultBackend) caPathExists() (bool, error) {
	path := filepath.Join(p.Path(), "cert", "ca")

	s, err := p.kubernetes.vaultClient.Logical().Read(path)
	if err != nil {
		return false, fmt.Errorf("error reading ca path '%s': %v", path, err)
	}

	if s == nil {
		return false, nil
	}
	if val, ok := s.Data["certificate"]; !ok || val == "" {
		return false, nil
	}

	return true, nil
}

func (p *PKIVaultBackend) WriteRole(role *pkiRole) error {
	_, err := p.kubernetes.vaultClient.Logical().Write(p.rolePath(role.Name), role.Data)
	if err != nil {
		return fmt.Errorf("error writting role '%s' to '%s': %v", role.Name, p.Path(), err)
	}

	return nil
}

func (p *PKIVaultBackend) DeleteRole(role *pkiRole) error {
	s, err := p.kubernetes.vaultClient.Logical().Read(p.rolePath(role.Name))
	if err != nil || s == nil || s.Data == nil {
		return nil
	}

	_, err = p.kubernetes.vaultClient.Logical().Delete(p.rolePath(role.Name))
	if err != nil {
		return fmt.Errorf("error deleting role '%s' to '%s': %v", role.Name, p.Path(), err)
	}

	return nil
}

func (p *PKIVaultBackend) ReadRole(role *pkiRole) (*vault.Secret, error) {
	path := filepath.Join(p.Path(), "roles", role.Name)

	secret, err := p.kubernetes.vaultClient.Logical().Read(path)
	if err != nil {
		return nil, fmt.Errorf("error reading role '%s' to '%s': %v", role.Name, p.Path(), err)
	}

	return secret, nil
}

func (p *PKIVaultBackend) Path() string {
	return filepath.Join(p.kubernetes.Path(), p.Type(), p.pkiName)
}

func (p *PKIVaultBackend) getMountConfigInput() vault.MountConfigInput {
	return vault.MountConfigInput{
		DefaultLeaseTTL: p.getDefaultLeaseTTL(),
		MaxLeaseTTL:     p.getMaxLeaseTTL(),
	}
}

func (p *PKIVaultBackend) getDefaultLeaseTTL() string {
	return fmt.Sprintf("%ds", int(p.DefaultLeaseTTL.Seconds()))
}

func (p *PKIVaultBackend) getMaxLeaseTTL() string {
	return fmt.Sprintf("%ds", int(p.MaxLeaseTTL.Seconds()))
}

func (p *PKIVaultBackend) getTokenPolicyExists(name string) (bool, error) {
	policy, err := p.kubernetes.vaultClient.Sys().GetPolicy(name)
	if err != nil {
		return false, err
	}

	if policy == "" {
		p.Log.Debugf("Policy Not Found: %s", name)
		return false, nil
	}

	p.Log.Debugf("Policy Found: %s", name)

	return true, nil
}

func (p *PKIVaultBackend) caGenPath() string {
	return filepath.Join(p.Path(), "root", "generate", "internal")
}

// Type is the sting key of the vault backend type
func (p *PKIVaultBackend) Type() string {
	return "pki"
}

func (p *PKIVaultBackend) Name() string {
	return p.pkiName
}

func (p *PKIVaultBackend) rolePath(role string) string {
	return filepath.Join(p.Path(), "roles", role)
}
