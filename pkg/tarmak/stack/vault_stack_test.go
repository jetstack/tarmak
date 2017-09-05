package stack_test

import (
	"testing"

	"github.com/hashicorp/vault/http"
	"github.com/hashicorp/vault/vault"
	"github.com/jetstack/tarmak/pkg/tarmak/stack"
)

func TestVaultTunnel(t *testing.T) {
	core, _, _ := vault.TestCoreUnsealed(t)
	ln1, addr1 := http.TestServer(t, core)
	defer ln1.Close()
	ln2, addr2 := http.TestServer(t, core)
	defer ln2.Close()
	st := &stack.Stack{}
	st.SetOutput(
		map[string]interface{}{
			"instance_fqdns": []interface{}{addr1, addr2},
			"vault_ca":       "",
		},
	)
	vs := &stack.VaultStack{
		Stack: st,
	}
	tunnel, err := vs.VaultTunnel()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(tunnel)
}
