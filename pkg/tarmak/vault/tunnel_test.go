// Copyright Jetstack Ltd. See LICENSE for details.
package vault_test

import (
	"crypto/x509"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/jetstack/tarmak/pkg/tarmak/interfaces"
	"github.com/jetstack/tarmak/pkg/tarmak/vault"
)

type FakeTunnel struct {
	bindAddress string
	port        string
}

func (ft *FakeTunnel) Port() string {
	return ft.port
}

func (ft *FakeTunnel) BindAddress() string {
	return ft.bindAddress
}

func (ft *FakeTunnel) Start() error {
	return nil
}

func (ft *FakeTunnel) Stop() {
	return
}

func (ft *FakeTunnel) Done() <-chan struct{} {
	return nil
}

var _ interfaces.Tunnel = &FakeTunnel{}

func TestVaultTunnel(t *testing.T) {
	ts := httptest.NewTLSServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				fmt.Fprintln(
					w,
					`{"initialized":true,
					  "sealed":false,
					  "standby":true,
					  "server_time_utc":1504683364,
					  "version":"0.7.3",
					  "cluster_name":"vault-test",
					  "cluster_id":"test"}`,
				)
			},
		),
	)

	u, err := url.Parse(ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	tunnel := &FakeTunnel{
		bindAddress: u.Hostname(),
		port:        u.Port(),
	}
	fqdn := "host1.example.com"
	vaultCA := x509.NewCertPool()
	vaultCACert, err := x509.ParseCertificate(ts.TLS.Certificates[0].Certificate[0])
	if err != nil {
		t.Fatal(err)
	}
	vaultCA.AddCert(vaultCACert)
	tun, err := vault.NewTunnel(tunnel, fqdn, vaultCA)
	if err != nil {
		t.Fatal(err)
	}
	c := tun.VaultClient()
	l, err := c.Sys().Health()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%#v", l)
	if !l.Standby {
		t.Error("'Standby' unexpectedly false")
	}
	if l.Sealed {
		t.Error("'Sealed' unexpectedly true")
	}
}
