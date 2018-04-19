// Copyright Jetstack Ltd. See LICENSE for details.
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

	vault "github.com/hashicorp/vault/api"
	"github.com/sirupsen/logrus"
)

type Generic struct {
	kubernetes *Kubernetes
	initTokens map[string]string

	Log *logrus.Entry
}

func (g *Generic) Ensure() error {
	err := g.GenerateSecretsMount()
	return err
}

func (g *Generic) Path() string {
	return filepath.Join(g.kubernetes.Path(), "secrets")
}

func (g *Generic) GenerateSecretsMount() error {
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
				Type:        "generic",
			},
		)

		if err != nil {
			return fmt.Errorf("error creating mount: %v", err)
		}

		g.Log.Infof("Mounted secrets: '%s'", g.Path())
	}

	rsaKeyPath := filepath.Join(g.Path(), "service-accounts")
	if secret, err := g.kubernetes.vaultClient.Logical().Read(rsaKeyPath); err != nil {
		return fmt.Errorf("error checking for secret %s: %v", rsaKeyPath, err)
	} else if secret == nil {
		err = g.writeNewRSAKey(rsaKeyPath, 4096)
		if err != nil {
			return fmt.Errorf("error creating rsa key at %s: %v", rsaKeyPath, err)
		}
	}

	return nil
}

func (g *Generic) writeNewRSAKey(secretPath string, bitSize int) error {
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

func (g *Generic) InitToken(name, role string, policies []string, expectedToken string) (string, error) {
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

func (g *Generic) initTokenPath(role string) string {
	return filepath.Join(g.Path(), fmt.Sprintf("init_token_%s", role))
}

func (g *Generic) InitTokenStore(role string) (token string, err error) {
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

func (g *Generic) revokeToken(token, path, role string) error {
	err := g.kubernetes.vaultClient.Auth().Token().RevokeOrphan(token)
	if err != nil {
		return fmt.Errorf("failed to revoke init token at path: %s", path)
	}

	g.Log.Infof("Revoked Token '%s': '%s'", role, token)

	return nil
}

func (g *Generic) SetInitTokenStore(role string, token string) error {
	path := g.initTokenPath(role)

	data := map[string]interface{}{
		"init_token": token,
	}
	_, err := g.kubernetes.vaultClient.Logical().Write(path, data)
	if err != nil {
		return fmt.Errorf("error writting init token at path:  %v", path)
	}

	g.Log.Infof("Init token written for '%s' at '%s'", role, path)

	return nil
}
