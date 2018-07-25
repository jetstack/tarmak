// Copyright Jetstack Ltd. See LICENSE for details.
package kubernetes

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/hashicorp/go-multierror"
	vault "github.com/hashicorp/vault/api"
	"github.com/sirupsen/logrus"
)

type GenericVaultBackend struct {
	kubernetes *Kubernetes
	initTokens map[string]string

	Log *logrus.Entry
}

func (g *GenericVaultBackend) Ensure() error {
	mount, err := GetMountByPath(g.kubernetes.vaultClient, g.Path())
	if err != nil {
		return err
	}

	if mount == nil {
		g.Log.Debugf("No secrects mount found for: %s", g.Path())
		err = g.kubernetes.vaultClient.Sys().Mount(
			g.Path(),
			&vault.MountInput{
				Description: "Kubernetes " + g.kubernetes.clusterID + " secrets",
				Type:        g.Type(),
			},
		)

		if err != nil {
			return fmt.Errorf("error creating mount: %v", err)
		}

		g.Log.Infof("Mounted secrets: '%s'", g.Path())
	}

	rsaKeyPath := g.ServiceAccountsPath()
	if secret, err := g.kubernetes.vaultClient.Logical().Read(rsaKeyPath); err != nil {
		return fmt.Errorf("error checking for secret %s: %v", rsaKeyPath, err)
	} else if secret == nil {
		err = g.writeNewRSAKey(rsaKeyPath, 4096)
		if err != nil {
			return fmt.Errorf("error creating rsa key at %s: %v", rsaKeyPath, err)
		}
	}

	if secret, err := g.kubernetes.vaultClient.Logical().Read(g.EncryptionConfigPath()); err != nil {
		return fmt.Errorf("error checking for secret %s: %v", g.EncryptionConfigPath(), err)
	} else if secret == nil {
		err = g.writeNewEncryptionConfig(g.EncryptionConfigPath())
		if err != nil {
			return fmt.Errorf("error creating encryption config at %s: %v", g.EncryptionConfigPath(), err)
		}
	}

	return nil
}

func (g *GenericVaultBackend) EnsureDryRun() (bool, error) {
	mount, err := GetMountByPath(g.kubernetes.vaultClient, g.Path())
	if err != nil {
		return false, err
	}

	if mount == nil || mount.Type != g.Type() {
		return true, nil
	}

	if secret, err := g.kubernetes.vaultClient.Logical().Read(g.ServiceAccountsPath()); err != nil {
		return false, fmt.Errorf("error checking for secret %s: %v", g.ServiceAccountsPath(), err)
	} else if secret == nil {
		return true, nil
	}

	if secret, err := g.kubernetes.vaultClient.Logical().Read(g.EncryptionConfigPath()); err != nil {
		return false, fmt.Errorf("error checking for secret %s: %v", g.EncryptionConfigPath(), err)
	} else if secret == nil {
		return true, nil
	}

	return false, nil
}

func (g *GenericVaultBackend) Delete() error {
	var result *multierror.Error

	if err := g.deleteSecret(g.ServiceAccountsPath()); err != nil {
		result = multierror.Append(result, err)
	}

	if err := g.deleteSecret(g.EncryptionConfigPath()); err != nil {
		result = multierror.Append(result, err)
	}

	if err := g.unMount(); err != nil {
		result = multierror.Append(result, err)
	}

	return result.ErrorOrNil()
}

func (g *GenericVaultBackend) Path() string {
	return filepath.Join(g.kubernetes.Path(), "secrets")
}

func (g *GenericVaultBackend) unMount() error {
	if err := g.kubernetes.vaultClient.Sys().Unmount(g.Path()); err != nil {
		return fmt.Errorf("failed to unmount secrects mount: %v", err)
	}

	return nil
}

func (g *GenericVaultBackend) writeNewRSAKey(secretPath string, bitSize int) error {
	reader := rand.Reader
	key, err := rsa.GenerateKey(reader, bitSize)
	if err != nil {
		return fmt.Errorf("error generating rsa key: %v", err)
	}

	var privateKey = &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	}

	var buf bytes.Buffer
	writer := bufio.NewWriter(&buf)
	err = pem.Encode(writer, privateKey)
	if err != nil {
		return fmt.Errorf("error encoding rsa key in PEM: %v", err)
	}
	if err := writer.Flush(); err != nil {
		return fmt.Errorf("error flushing buffer: %v", err)
	}

	writeData := map[string]interface{}{
		"key": buf.String(),
	}

	_, err = g.kubernetes.vaultClient.Logical().Write(secretPath, writeData)
	if err != nil {
		return fmt.Errorf("error writting key to secrets: %v", err)
	}

	g.Log.Infof("Key written to secrets '%s'", secretPath)

	return nil
}

func (g *GenericVaultBackend) writeNewEncryptionConfig(secretPath string) error {
	encryptionConfig := `kind: EncryptionConfig
apiVersion: v1
resources:
  - resources:
    - secrets
    - configmaps
    providers:
    - aescbc:
        keys:
        - name: key1
          secret: SECRET
    - identity: {}
`

	secret := make([]byte, 32)

	_, err := rand.Read(secret)
	if err != nil {
		return fmt.Errorf("error generating secret: %v", err)
	}

	writeData := map[string]interface{}{
		"content": strings.Replace(encryptionConfig, "SECRET", base64.StdEncoding.EncodeToString(secret), 1),
	}

	_, err = g.kubernetes.vaultClient.Logical().Write(secretPath, writeData)
	if err != nil {
		return fmt.Errorf("error writing key to secrets: %v", err)
	}

	g.Log.Infof("Key written to secrets '%s'", secretPath)

	return nil
}

func (g *GenericVaultBackend) deleteSecret(secretPath string) error {
	s, err := g.kubernetes.vaultClient.Logical().Read(secretPath)
	if err != nil || s == nil || s.Data == nil {
		return nil
	}

	_, err = g.kubernetes.vaultClient.Logical().Delete(secretPath)
	if err != nil {
		return fmt.Errorf("error deleting key from secrets: %v", err)
	}

	return nil
}

func (g *GenericVaultBackend) InitToken(name, role string, policies []string, expectedToken string) (string, error) {
	path := g.initTokenPath(role)

	if secret, err := g.kubernetes.vaultClient.Logical().Read(path); err != nil {
		return "", fmt.Errorf("error checking for secret %s: %v", path, err)
	} else if secret != nil {
		key := "init_token"
		token, ok := secret.Data[key]
		if !ok {
			return "", fmt.Errorf("error secret %s doesn't contain a key '%s'", path, key)
		}

		tokenStr, ok := token.(string)
		if !ok {
			return "", fmt.Errorf("error secret %s key '%s' has wrong type: %T", path, key, token)
		}

		return tokenStr, nil
	}

	// we have to create a new token
	tokenRequest := &vault.TokenCreateRequest{
		ID:          expectedToken,
		DisplayName: name,
		TTL:         fmt.Sprintf("%ds", int(g.kubernetes.MaxValidityInitTokens.Seconds())),
		Period:      fmt.Sprintf("%ds", int(g.kubernetes.MaxValidityInitTokens.Seconds())),
		Policies:    policies,
	}

	token, err := g.kubernetes.vaultClient.Auth().Token().CreateOrphan(tokenRequest)
	if err != nil {
		return "", fmt.Errorf("failed to create init token: %v", err)
	}

	err = g.SetInitTokenStore(role, token.Auth.ClientToken)
	if err != nil {
		return "", fmt.Errorf("failed to store init token in '%s': %v", path, err)
	}

	return token.Auth.ClientToken, nil
}

func (g *GenericVaultBackend) initTokenPath(role string) string {
	return filepath.Join(g.Path(), fmt.Sprintf("init_token_%s", role))
}

func (g *GenericVaultBackend) InitTokenStore(role string) (token string, err error) {
	path := g.initTokenPath(role)

	s, err := g.kubernetes.vaultClient.Logical().Read(path)
	if err != nil {
		return "", fmt.Errorf("failed to read init token: %v", err)
	}
	if s == nil {
		return "", nil
	}

	dat, ok := s.Data["init_token"]
	if !ok {
		return "", fmt.Errorf("failed to find init token data at '%s': %v", path, err)
	}
	token, ok = dat.(string)
	if !ok {
		return "", fmt.Errorf("failed to convert token data to string: %v", err)
	}

	return token, nil
}

func (g *GenericVaultBackend) revokeToken(token, path, role string) error {
	err := g.kubernetes.vaultClient.Auth().Token().RevokeOrphan(token)
	if err != nil {
		return fmt.Errorf("failed to revoke init token at path '%s': %v", path, err)
	}

	g.Log.Infof("Revoked Token '%s': '%s'", role, token)

	return nil
}

func (g *GenericVaultBackend) SetInitTokenStore(role string, token string) error {
	path := g.initTokenPath(role)

	data := map[string]interface{}{
		"init_token": token,
	}
	_, err := g.kubernetes.vaultClient.Logical().Write(path, data)
	if err != nil {
		return fmt.Errorf("error writting init token at path '%s': %v", path, err)
	}

	g.Log.Infof("Init token written for '%s' at '%s'", role, path)

	return nil
}

func (g *GenericVaultBackend) DeleteInitTokenStore(role string) error {
	path := g.initTokenPath(role)

	_, err := g.kubernetes.vaultClient.Logical().Delete(path)
	if err != nil {
		return fmt.Errorf("error deleting init token at path '%s': %v", path, err)
	}

	return nil
}

func (g *GenericVaultBackend) Type() string {
	return "generic"
}

func (g *GenericVaultBackend) Name() string {
	return "secrets"
}

// ServiceAccountsPath is the vault path for the service-accounts certificate content
func (g *GenericVaultBackend) ServiceAccountsPath() string {
	return filepath.Join(g.Path(), "service-accounts")
}

// EncryptionConfigPath is the vault path for the kubernetes encryption config file content
func (g *GenericVaultBackend) EncryptionConfigPath() string {
	return filepath.Join(g.Path(), "encryption-config")
}
