package kubernetes

import (
	"strings"
	"testing"
)

func TestIsValidClusterID(t *testing.T) {
	var err error

	err = isValidClusterID("valid-cluster")
	if err != nil {
		t.Error("unexpected an error: ", err)
	}

	err = isValidClusterID("valid-cluster01")
	if err != nil {
		t.Error("unexpected an error: ", err)
	}

	err = isValidClusterID("")
	if err == nil {
		t.Error("expected an error")
	} else if msg := "Invalid cluster ID"; !strings.Contains(err.Error(), msg) {
		t.Errorf("error '%v' should contain '%s'", err, msg)
	}

	err = isValidClusterID("invalid.cluster")
	if err == nil {
		t.Error("expected an error")
	} else if msg := "Invalid cluster ID"; !strings.Contains(err.Error(), msg) {
		t.Errorf("error '%v' should contain '%s'", err, msg)
	}
}

func TestKubernetes_Ensure(t *testing.T) {
	vault := NewFakeVault(t)
	defer vault.Finish()
	k := vault.Kubernetes()

	vault.Ensure()

	err := k.Ensure()
	if err != nil {
		t.Fatalf("error ensuring: %v", err)
		return
	}
}

func TestKubernetes_NewToken_Role(t *testing.T) {
	vault := NewFakeVault(t)
	defer vault.Finish()
	k := vault.Kubernetes()

	vault.NewToken()

	adminRole := k.k8sAdminRole()

	err := k.kubernetesPKI.WriteRole(adminRole)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
		return
	}

	kubeSchedulerRole := k.k8sComponentRole("kube-scheduler")

	err = k.kubernetesPKI.WriteRole(kubeSchedulerRole)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
