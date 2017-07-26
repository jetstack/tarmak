package kubernetes

import (
	"testing"
	"time"
)

func TestPKI_Ensure(t *testing.T) {
	vault := NewFakeVault(t)
	defer vault.Finish()
	k := vault.Kubernetes()

	vault.PKIEnsure()

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
	if exp, act := "test-cluster-inside/generic", k.secretsGeneric.Path(); exp != act {
		t.Errorf("unexpected value, exp=%s got=%s", exp, act)
		return
	}

	k.etcdKubernetesPKI.DefaultLeaseTTL = time.Hour * 0
	k.etcdOverlayPKI.MaxLeaseTTL = time.Hour * 0
	k.kubernetesPKI.DefaultLeaseTTL = time.Hour * 0
	if err := k.Ensure(); err != nil {
		t.Error("unexpected error", err)
		return
	}

	policy_name := k.clusterID + "/" + "master"

	exists, err := k.etcdKubernetesPKI.getTokenPolicyExists(policy_name)
	if err != nil {
		t.Error("Error finding policy: ", err)
		return
	}
	if exists {
		t.Error("Policy Found - it should not be")
		return
	}

	policy := k.masterPolicy()

	err = k.WritePolicy(policy)
	if err != nil {
		t.Error("Error writting policy: ", err)
		return
	}

	exists, err = k.etcdKubernetesPKI.getTokenPolicyExists(policy_name)
	if err != nil {
		t.Error("Error finding policy: ", err)
		return
	}
	if !exists {
		t.Error("Policy not found")
		return
	}

	//pkiWrongType := NewPKI(k, "wrong-type-pki")

	//err = k.vaultClient.Sys().Mount(
	//	k.Path()+"/pki/"+"wrong-type-pki",
	//	&vault_testing.MountInput{
	//		Description: "Kubernetes " + k.clusterID + "/" + "wrong-type-pki" + " CA",
	//		Type:        "generic",
	//		Config:      k.etcdKubernetesPKI.getMountConfigInput(),
	//	},
	//)
	//if err != nil {
	//	t.Error("Error Mounting: ", err)
	//	return
	//}

	//err = pkiWrongType.Ensure()
	//if err == nil {
	//	t.Error("Should have error from wrong type")
	//	return
	//}

}
