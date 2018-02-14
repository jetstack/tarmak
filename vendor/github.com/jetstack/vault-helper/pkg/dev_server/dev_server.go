package dev_server

import (
	"github.com/sirupsen/logrus"

	"github.com/jetstack/vault-helper/pkg/kubernetes"
	"github.com/jetstack/vault-helper/pkg/testing/vault_dev"
)

const FlagWaitSignal = "wait-signal"
const FlagPortNumber = "port"

type DevVault struct {
	Vault      *vault_dev.VaultDev
	Kubernetes *kubernetes.Kubernetes
	Log        *logrus.Entry
}

func New(logger *logrus.Entry) *DevVault {
	vault := vault_dev.New()

	v := &DevVault{
		Vault: vault,
		Log:   logger,
	}

	return v
}
