package instanceToken

import (
	"path/filepath"

	"github.com/Sirupsen/logrus"
	vault "github.com/hashicorp/vault/api"
)

const FlagInitRole = "init-role"
const FlagConfigPath = "config-path"

type InstanceToken struct {
	token           string
	initRole        string
	vaultConfigPath string

	Log         *logrus.Entry
	vaultClient *vault.Client
}

func (i *InstanceToken) SetInitRole(initRole string) {
	i.initRole = initRole
}

func (i *InstanceToken) InitRole() (initRole string) {
	return i.initRole
}

func (i *InstanceToken) SetToken(token string) {
	i.token = token
}

func (i *InstanceToken) Token() (token string) {
	return i.token
}

func (i *InstanceToken) SetVaultConfigPath(path string) {
	i.vaultConfigPath = path
}

func (i *InstanceToken) VaultConfigPath() (path string) {
	return i.vaultConfigPath
}

func (i *InstanceToken) TokenFilePath() (path string) {
	return filepath.Join(i.VaultConfigPath(), "token")
}
func (i *InstanceToken) InitTokenFilePath() (path string) {
	return filepath.Join(i.VaultConfigPath(), "init-token")
}

func (i *InstanceToken) VaultClient() (vaultClient *vault.Client) {
	return i.vaultClient
}

func New(vaultClient *vault.Client, logger *logrus.Entry) *InstanceToken {
	i := &InstanceToken{}

	if vaultClient != nil {
		i.vaultClient = vaultClient
	}

	if logger != nil {
		i.Log = logger
	}

	return i
}
