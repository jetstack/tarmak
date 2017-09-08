package stack_test

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"

	"github.com/jetstack/tarmak/pkg/tarmak/interfaces"
	"github.com/jetstack/tarmak/pkg/tarmak/stack"
)

type FakeTunnel struct {
	bindAddress string
	port        int
}

func (ft *FakeTunnel) Port() int {
	return ft.port
}

func (ft *FakeTunnel) BindAddress() string {
	return ft.bindAddress
}

func (ft *FakeTunnel) Start() error {
	return nil
}

func (ft *FakeTunnel) Stop() error {
	return nil
}

var _ interfaces.Tunnel = &FakeTunnel{}

func ServerAndClientForTest(
	t *testing.T, handler http.Handler,
) (*httptest.Server, *http.Client) {
	ts := httptest.NewTLSServer(handler)
	// XXX: More verbose than it needs to be.
	// See https://github.com/golang/go/issues/18411
	cert, err := x509.ParseCertificate(ts.TLS.Certificates[0].Certificate[0])
	if err != nil {
		t.Fatal(err)
	}

	certpool := x509.NewCertPool()
	certpool.AddCert(cert)
	tc := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs: certpool,
			},
		},
	}
	return ts, tc
}

func TestVaultTunnel(t *testing.T) {
	ts, _ := ServerAndClientForTest(
		t,
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
	port, err := strconv.Atoi(u.Port())
	if err != nil {
		t.Fatal(err)
	}
	tunnel := &FakeTunnel{
		bindAddress: u.Hostname(),
		port:        port,
	}
	fqdn := "host1.example.com"
	vaultCA := ts.TLS.Certificates[0].Certificate[0]
	tun, err := stack.NewVaultTunnel(tunnel, fqdn, vaultCA)
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
