package kubernetes_pki_test

import (
	"testing"
	"time"

	vault "github.com/hashicorp/vault/api"

	"github.com/jetstack-experimental/vault-helper/pkg/kubernetes_pki"
)

func TestPKI_Ensure(t *testing.T) {

	testPath := "test-vault-helper-pki-ensure"

	vaultClient, err := vault.NewClient(nil)
	if err != nil {
		t.Skip("Unable to create vault client, skipping integration tests: ", err)
	}

	_, err = vaultClient.Auth().Token().LookupSelf()
	if err != nil {
		t.Skip("Unable to lookup my token, skipping integration tests: ", err)
	}

	// should create non existing mount
	pki := kubernetes_pki.NewPKI(vaultClient, testPath)
	firstMaxLeaseTTL := 120 * time.Minute
	pki.MaxLeaseTTL = firstMaxLeaseTTL
	firstDefaultLeaseTTL := 60 * time.Minute
	pki.DefaultLeaseTTL = firstDefaultLeaseTTL
	pki.Description = "first description"
	if err = pki.Ensure(); err != nil {
		t.Error("Unexpected error:", err)
	}

	// should max TTL description
	secondMaxLeaseTTL := 120 * time.Second
	pki.MaxLeaseTTL = secondMaxLeaseTTL
	secondDefaultLeaseTTL := 60 * time.Second
	pki.DefaultLeaseTTL = secondDefaultLeaseTTL
	if err = pki.Ensure(); err != nil {
		t.Error("Unexpected error:", err)
	}
	mount, err := kubernetes_pki.GetMountByPath(vaultClient, testPath)
	if err != nil {
		t.Error("Unexpected error:", err)
	}
	if exp, act := int(secondMaxLeaseTTL.Seconds()), mount.Config.MaxLeaseTTL; exp != act {
		t.Errorf("Did not update description: exp=%d act=%d", exp, act)
	}
}
