// Copyright Jetstack Ltd. See LICENSE for details.
package vault

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"

	cleanhttp "github.com/hashicorp/go-cleanhttp"
	vault "github.com/hashicorp/vault/api"

	"github.com/jetstack/tarmak/pkg/tarmak/interfaces"
)

type vaultTunnel struct {
	tunnel      interfaces.Tunnel
	tunnelError error
	client      *vault.Client
	fqdn        string
}

var _ interfaces.Tunnel = &vaultTunnel{}

func NewTunnel(
	tunnel interfaces.Tunnel, fqdn string, vaultCA *x509.CertPool,
) (*vaultTunnel, error) {
	httpTransport := cleanhttp.DefaultTransport()
	httpTransport.TLSClientConfig = &tls.Config{
		RootCAs: vaultCA,
	}
	httpClient := cleanhttp.DefaultClient()
	httpClient.Transport = httpTransport
	config := vault.DefaultConfig()
	config.HttpClient = httpClient
	vaultClient, err := vault.NewClient(config)
	if err != nil {
		return &vaultTunnel{}, err
	}
	err = vaultClient.SetAddress(
		fmt.Sprintf(
			"https://%s:%d", tunnel.BindAddress(), tunnel.Port(),
		),
	)
	if err != nil {
		return &vaultTunnel{}, err
	}
	return &vaultTunnel{
		tunnel: tunnel,
		client: vaultClient,
		fqdn:   fqdn,
	}, nil
}

func (v *vaultTunnel) FQDN() string {
	return v.fqdn
}
func (v *vaultTunnel) Start() error {

	if err := v.tunnel.Start(); err != nil {
		v.tunnelError = err
		return err
	}

	return nil
}

func (v *vaultTunnel) Stop() {
	v.tunnel.Stop()
}

func (v *vaultTunnel) Port() int {
	return v.tunnel.Port()
}

func (v *vaultTunnel) BindAddress() string {
	return v.tunnel.BindAddress()
}

func (v *vaultTunnel) VaultClient() *vault.Client {
	return v.client
}

func (v *vaultTunnel) Status() int {
	if v.tunnelError != nil {
		return VaultStateErr
	}

	initStatus, err := v.client.Sys().InitStatus()
	if err != nil {
		return VaultStateErr
	}

	if !initStatus {
		return VaultStateUnintialised
	}

	sealStatus, err := v.client.Sys().SealStatus()
	if err != nil {
		return VaultStateErr
	}

	if sealStatus.Sealed {
		return VaultStateSealed
	}
	return VaultStateUnsealed
}
