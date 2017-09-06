package stack_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"

	vault "github.com/hashicorp/vault/api"

	"github.com/jetstack/tarmak/pkg/tarmak/interfaces"
	"github.com/jetstack/tarmak/pkg/tarmak/stack"
)

type FakeTunnel struct {
	port int
}

func (ft *FakeTunnel) Port() int {
	return ft.port
}

func (ft *FakeTunnel) Start() error {
	return nil
}

func (ft *FakeTunnel) Stop() error {
	return nil
}

var _ interfaces.Tunnel = &FakeTunnel{}

func TestVaultTunnel(t *testing.T) {
	s := httptest.NewTLSServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusTooManyRequests)
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
	defer s.Close()
	u, err := url.Parse(s.URL)
	if err != nil {
		t.Fatal(err)
	}
	port, err := strconv.Atoi(u.Port())
	if err != nil {
		t.Fatal(err)
	}
	tunnel := &FakeTunnel{port: port}
	client, err := vault.NewClient(nil)
	if err != nil {
		t.Fatal(err)
	}
	fqdn := "host1.example.com"
	tun := stack.NewVaultTunnel(tunnel, client, fqdn)
	if err != nil {
		t.Fatal(err)
	}
	c := tun.VaultClient()
	l, err := c.Sys().Leader()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(l)
}
