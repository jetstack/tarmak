package kubernetes_pki

import (
	"time"

	"github.com/Sirupsen/logrus"
	vault "github.com/hashicorp/vault/api"
)

type KubernetesPKI struct {
	prefix      string
	vaultClient vault.Client
	log         *logrus.Entry

	// validity for vault
	MaxValidityComponents time.Duration
	MaxValidityAdmin      time.Duration
	MaxValidityCA         time.Duration
}

func New(prefix string, vaultClient *vault.Client) *KubernetesPKI {
	return &KubernetesPKI{
		prefix: prefix,
	}
}
