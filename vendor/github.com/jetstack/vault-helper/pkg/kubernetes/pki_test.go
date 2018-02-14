package kubernetes

import (
	"testing"
	"time"

	vault "github.com/hashicorp/vault/api"
	"github.com/sirupsen/logrus"
)

func TestPKI_Ensure(t *testing.T) {
	fakeVault := NewFakeVault(t)
	defer fakeVault.Finish()
	k := fakeVault.Kubernetes()

	fakeVault.PKIEnsure()

	if exp, act := "test-cluster-inside/pki/etcd-k8s", k.etcdKubernetesPKI.Path(); exp != act {
		t.Errorf("unexpected value, exp=%s got=%s", exp, act)
		return
	}
	if exp, act := "test-cluster-inside/pki/etcd-overlay", k.etcdOverlayPKI.Path(); exp != act {
		t.Errorf("unexpected value, exp=%s got=%s", exp, act)
		return
	}
	if exp, act := "test-cluster-inside/pki/k8s", k.kubernetesPKI.Path(); exp != act {
		t.Errorf("unexpected value, exp=%s got=%s", exp, act)
		return
	}
	if exp, act := "test-cluster-inside/secrets", k.secretsGeneric.Path(); exp != act {
		t.Errorf("unexpected value, exp=%s got=%s", exp, act)
		return
	}

	k.etcdKubernetesPKI.DefaultLeaseTTL = time.Hour * 0
	k.etcdOverlayPKI.MaxLeaseTTL = time.Hour * 0
	k.kubernetesPKI.DefaultLeaseTTL = time.Hour * 0
	if err := k.Ensure(); err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	policy_name := k.clusterID + "/" + "master"

	exists, err := k.etcdKubernetesPKI.getTokenPolicyExists(policy_name)
	if err != nil {
		t.Errorf("failed to find policy: %v", err)
		return
	}
	if exists {
		t.Error("unexpected policy found")
		return
	}

	policy := k.masterPolicy()

	err = k.WritePolicy(policy)
	if err != nil {
		t.Errorf("failed to write policy: %v", err)
		return
	}

	exists, err = k.etcdKubernetesPKI.getTokenPolicyExists(policy_name)
	if err != nil {
		t.Errorf("faileds to find policy: %v", err)
		return
	}
	if !exists {
		t.Error("policy not found")
		return
	}

	pkiWrongType := NewPKI(k, "wrong-type-pki", logrus.NewEntry(logrus.New()))

	err = k.vaultClient.Sys().Mount(
		k.Path()+"/pki/"+"wrong-type-pki",
		&vault.MountInput{
			Description: "Kubernetes " + k.clusterID + "/" + "wrong-type-pki" + " CA",
			Type:        "generic",
			Config:      k.etcdKubernetesPKI.getMountConfigInput(),
		},
	)
	if err != nil {
		t.Errorf("failed to mount: %v", err)
		return
	}

	err = pkiWrongType.Ensure()
	if err == nil {
		t.Errorf("expected an error from wrong pki type")
		return
	}
}
