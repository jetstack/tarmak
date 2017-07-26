package kubernetes_pki

import (
	"fmt"
	"path"
	"time"

	"github.com/Sirupsen/logrus"
	vault "github.com/hashicorp/vault/api"
)

type PKI struct {
	vaultClient *vault.Client
	path        string
	log         *logrus.Entry

	Description string

	MaxLeaseTTL     time.Duration
	DefaultLeaseTTL time.Duration
}

func NewPKI(vaultClient *vault.Client, path string) *PKI {
	return &PKI{
		vaultClient: vaultClient,
		path:        path,
		log:         logrus.WithField("type", "pki").WithField("path", path),

		MaxLeaseTTL:     time.Hour * 24 * 60,
		DefaultLeaseTTL: time.Hour * 24 * 30,
	}
}

func GetMountByPath(vaultClient *vault.Client, mountPath string) (*vault.MountOutput, error) {
	mounts, err := vaultClient.Sys().ListMounts()
	if err != nil {
		return nil, fmt.Errorf("error listing mounts: %s", err)
	}

	var mount *vault.MountOutput
	for key, _ := range mounts {
		if path.Clean(key) == path.Clean(mountPath) {
			mount = mounts[key]
			break
		}
	}

	return mount, nil
}

func (p *PKI) getDefaultLeaseTTL() string {
	return fmt.Sprintf("%d", int(p.DefaultLeaseTTL.Seconds()))
}

func (p *PKI) getMaxLeaseTTL() string {
	return fmt.Sprintf("%d", int(p.MaxLeaseTTL.Seconds()))
}

func (p *PKI) getMountConfigInput() vault.MountConfigInput {
	return vault.MountConfigInput{
		DefaultLeaseTTL: p.getDefaultLeaseTTL(),
		MaxLeaseTTL:     p.getMaxLeaseTTL(),
	}
}

func (p *PKI) Ensure() error {

	mount, err := GetMountByPath(p.vaultClient, p.path)
	if err != nil {
		return err
	}

	if mount == nil {
		err := p.vaultClient.Sys().Mount(
			p.path,
			&vault.MountInput{
				Description: p.Description,
				Type:        "pki",
				Config:      p.getMountConfigInput(),
			},
		)
		if err != nil {
			return fmt.Errorf("error creating mount: %s", err)
		}
		p.log.Info("created mount")
		return nil
	} else {
		if mount.Type != "pki" {
			return fmt.Errorf("mount '%s' already existing with wrong type '%s'", p.path, mount.Type)
		}
		if mount.Description != p.Description {
			return fmt.Errorf("update description unimplemented")
		}
	}

	tuneMountRequired := false

	if mount.Config.DefaultLeaseTTL != int(p.DefaultLeaseTTL.Seconds()) {
		tuneMountRequired = true
	}
	if mount.Config.MaxLeaseTTL != int(p.MaxLeaseTTL.Seconds()) {
		tuneMountRequired = true
	}

	if tuneMountRequired {
		mountConfig := p.getMountConfigInput()
		err := p.vaultClient.Sys().TuneMount(p.path, mountConfig)
		if err != nil {
			return fmt.Errorf("error tuning mount config: %s", err.Error())
		}
		p.log.Infof("tuned mount config=%+v")
	}

	return nil

}
