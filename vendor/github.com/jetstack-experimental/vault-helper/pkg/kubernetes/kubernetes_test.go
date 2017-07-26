package kubernetes

import (
	"strings"
	"testing"
)

func TestIsValidClusterID(t *testing.T) {
	var err error

	err = isValidClusterID("valid-cluster")
	if err != nil {
		t.Error("unexpected an error: %s", err)
	}

	err = isValidClusterID("valid-cluster01")
	if err != nil {
		t.Error("unexpected an error: %s", err)
	}

	err = isValidClusterID("")
	if err == nil {
		t.Error("expected an error")
	} else if msg := "Invalid cluster ID"; !strings.Contains(err.Error(), msg) {
		t.Errorf("error '%s' should contain '%s'", err, msg)
	}

	err = isValidClusterID("invalid.cluster")
	if err == nil {
		t.Error("expected an error")
	} else if msg := "Invalid cluster ID"; !strings.Contains(err.Error(), msg) {
		t.Errorf("error '%s' should contain '%s'", err, msg)
	}

}

//go test -coverprofile=coverage.out
//  go tool cover -html=coverage.out

func TestKubernetes_Double_Ensure(t *testing.T) {
	vault := NewFakeVault(t)
	defer vault.Finish()
	k := vault.Kubernetes()

	vault.DoubleEnsure()

	err := k.Ensure()
	if err != nil {
		t.Error("error ensuring: ", err)
		return
	}

	err = k.Ensure()
	if err != nil {
		t.Error("error double ensuring: ", err)
		return
	}

}

func TestKubernetes_NewPolicy_Role(t *testing.T) {
	vault := NewFakeVault(t)
	defer vault.Finish()
	k := vault.Kubernetes()

	vault.NewPolicy()

	masterPolicy := k.masterPolicy()

	err := k.WritePolicy(masterPolicy)
	if err != nil {
		t.Error("unexpected error", err)
		return
	}
}

//func TestKubernetes_NewToken_Role(t *testing.T) {
//	vault := NewFakeVault(t)
//	defer vault.Finish()
//	k := vault.Kubernetes()
//
//	writeData := map[string]interface{}{
//		"use_csr_common_name": false,
//		"enforce_hostnames":   false,
//		"organization":        "system:masters",
//		"allowed_domains":     "admin",
//		"allow_bare_domains":  true,
//		"allow_localhost":     false,
//		"allow_subdomains":    false,
//		"allow_ip_sans":       false,
//		"server_flag":         false,
//		"client_flag":         true,
//		"max_ttl":             "140h",
//		"ttl":                 "140h",
//	}
//
//	adminRole := k.NewTokenRole("admin", writeData)
//
//	err := adminRole.WriteTokenRole()
//
//	if err != nil {
//		t.Error("unexpected error", err)
//		return
//	}
//
//	writeData = map[string]interface{}{
//		"use_csr_common_name": false,
//		"enforce_hostnames":   false,
//		"allowed_domains":     "kube-scheduler,system:kube-scheduler",
//		"allow_bare_domains":  true,
//		"allow_localhost":     false,
//		"allow_subdomains":    false,
//		"allow_ip_sans":       false,
//		"server_flag":         false,
//		"client_flag":         true,
//		"max_ttl":             "140h",
//		"ttl":                 "140h",
//	}
//
//	kubeSchedulerRole := k.NewTokenRole("kube-scheduler", writeData)
//
//	err = kubeSchedulerRole.WriteTokenRole()
//
//	if err != nil {
//		t.Error("unexpected error", err)
//		return
//	}
//
//}
