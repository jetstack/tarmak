package kubernetes

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"path/filepath"

	"github.com/Sirupsen/logrus"
	"github.com/google/uuid"
	vault "github.com/hashicorp/vault/api"
)

type Generic struct {
	kubernetes *Kubernetes
	initTokens map[string]string
}

func (g *Generic) Ensure() (bool, error) {
	err := g.GenerateSecretsMount()
	return false, err
}

func (g *Generic) Path() string {
	return filepath.Join(g.kubernetes.Path(), "secrets")
}

func randomUUID() string {
	uuID := uuid.New()
	return uuID.String()
}

func (g *Generic) GenerateSecretsMount() error {

	mount, err := GetMountByPath(g.kubernetes.vaultClient, g.Path())
	if err != nil {
		return err
	}

	if mount == nil {
		logrus.Debugf("No secrects mount found for: %s", g.Path())
		err = g.kubernetes.vaultClient.Sys().Mount(
			g.Path(),
			&vault.MountInput{
				Description: "Kubernetes " + g.kubernetes.clusterID + " secrets",
				Type:        "generic",
			},
		)

		if err != nil {
			return fmt.Errorf("error creating mount: %s", err)
		}

		logrus.Infof("Mounted secrets: '%s'", g.Path())
	}

	rsaKeyPath := filepath.Join(g.Path(), "service-accounts")
	if secret, err := g.kubernetes.vaultClient.Logical().Read(rsaKeyPath); err != nil {
		return fmt.Errorf("error checking for secret %s: %s", rsaKeyPath, err)
	} else if secret == nil {
		err = g.writeNewRSAKey(rsaKeyPath, 4096)
		if err != nil {
			return fmt.Errorf("error creating rsa key at %s: %s", rsaKeyPath, err)
		}
	}

	return nil
}

func (g *Generic) writeNewRSAKey(secretPath string, bitSize int) error {

	reader := rand.Reader
	key, err := rsa.GenerateKey(reader, bitSize)
	if err != nil {
		return fmt.Errorf("error generating rsa key: %s", err)
	}

	var privateKey = &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	}

	var buf bytes.Buffer
	writer := bufio.NewWriter(&buf)
	err = pem.Encode(writer, privateKey)
	if err != nil {
		return fmt.Errorf("error encoding rsa key in PEM: %s", err)
	}
	if err := writer.Flush(); err != nil {
		return fmt.Errorf("error flushing buffer: %s", err)
	}

	writeData := map[string]interface{}{
		"key": buf.String(),
	}

	_, err = g.kubernetes.vaultClient.Logical().Write(secretPath, writeData)

	if err != nil {
		return fmt.Errorf("error writting key to secrets: %s", err)
	}
	logrus.Infof("Key written to secrets '%s'", secretPath)

	return nil
}

func (g *Generic) InitToken(name, role string, policies []string) (string, bool, error) {
	path := filepath.Join(g.Path(), fmt.Sprintf("init_token_%s", role))

	if secret, err := g.kubernetes.vaultClient.Logical().Read(path); err != nil {
		return "", false, fmt.Errorf("error checking for secret %s: %s", path, err)
	} else if secret != nil {
		key := "init_token"
		token, ok := secret.Data[key]
		if !ok {
			return "", false, fmt.Errorf("error secret %s doesn't contain a key '%s'", path, key)
		}

		tokenStr, ok := token.(string)
		if !ok {
			return "", false, fmt.Errorf("error secret %s key '%s' has wrong type: %T", path, key, token)
		}

		return tokenStr, false, nil
	}

	// we have to create a new token
	tokenRequest := &vault.TokenCreateRequest{
		DisplayName: name,
		TTL:         fmt.Sprintf("%ds", int(g.kubernetes.MaxValidityInitTokens.Seconds())),
		Period:      fmt.Sprintf("%ds", int(g.kubernetes.MaxValidityInitTokens.Seconds())),
		Policies:    policies,
	}

	token, err := g.kubernetes.vaultClient.Auth().Token().CreateOrphan(tokenRequest)
	if err != nil {
		return "", false, fmt.Errorf("failed to create init token: %s", err)
	}

	dataStoreToken := map[string]interface{}{
		"init_token": token.Auth.ClientToken,
	}
	_, err = g.kubernetes.vaultClient.Logical().Write(path, dataStoreToken)
	if err != nil {
		return "", false, fmt.Errorf("failed to store init token in '%s': %s", path, err)
	}

	return token.Auth.ClientToken, true, nil

}
