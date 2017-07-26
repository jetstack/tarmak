package kubernetes_test

import (
	"testing"

	//"github.com/Sirupsen/logrus"
	"github.com/jetstack-experimental/vault-helper/pkg/kubernetes"
	"github.com/jetstack-experimental/vault-helper/pkg/testing/vault_dev"
)

func TestGeneric_Ensure(t *testing.T) {
	vault := vault_dev.New()
	if err := vault.Start(); err != nil {
		t.Skip("unable to initialise vault dev server for integration tests: ", err)
	}
	defer vault.Stop()

	k := kubernetes.New(vault.Client())
	k.SetClusterID("test-cluster")

	generic := k.NewGeneric()
	_, err := generic.Ensure()
	if err != nil {
		t.Error("unexpected error: ", err)
		return
	}

	_, err = generic.Ensure()
	if err != nil {
		t.Error("unexpected error: ", err)
		return
	}
}
