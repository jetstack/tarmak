package kubernetes

import (
	"testing"

	"github.com/golang/mock/gomock"
	vault "github.com/hashicorp/vault/api"
)

type fakeVault struct {
	ctrl *gomock.Controller

	fakeVault   *MockVault
	fakeSys     *MockVaultSys
	fakeLogical *MockVaultLogical
	fakeAuth    *MockVaultAuth
}

func NewFakeVault(t *testing.T) *fakeVault {
	ctrl := gomock.NewController(t)

	v := &fakeVault{
		ctrl: ctrl,

		fakeVault:   NewMockVault(ctrl),
		fakeSys:     NewMockVaultSys(ctrl),
		fakeLogical: NewMockVaultLogical(ctrl),
	}

	v.fakeVault.EXPECT().Sys().AnyTimes().Return(v.fakeSys)
	v.fakeVault.EXPECT().Logical().AnyTimes().Return(v.fakeLogical)
	v.fakeVault.EXPECT().Auth().AnyTimes().Return(v.fakeAuth)

	return v
}

func (v *fakeVault) Kubernetes() *Kubernetes {
	k := New(nil)
	k.SetClusterID("test-cluster-inside")
	k.vaultClient = v.fakeVault
	return k
}

func (v *fakeVault) Finish() {
	v.ctrl.Finish()
}

func (v *fakeVault) DoubleEnsure() {

	mountInput1 := &vault.MountInput{
		Description: "Kubernetes test-cluster-inside/etcd-k8s CA",
		Type:        "pki",
	}

	mountInput2 := &vault.MountInput{
		Description: "Kubernetes " + "test-cluster-inside" + "/" + "etcd-overlay" + " CA",
		Type:        "pki",
	}

	mountInput3 := &vault.MountInput{
		Description: "Kubernetes " + "test-cluster-inside" + "/" + "k8s" + " CA",
		Type:        "pki",
	}

	v.fakeSys.EXPECT().ListMounts().AnyTimes().Return(nil, nil)

	v.fakeSys.EXPECT().Mount("test-cluster-inside/pki/etcd-k8s", mountInput1).Times(2).Return(nil)
	v.fakeSys.EXPECT().Mount("test-cluster-inside/pki/etcd-overlay", mountInput2).Times(2).Return(nil)
	v.fakeSys.EXPECT().Mount("test-cluster-inside/pki/k8s", mountInput3).Times(2).Return(nil)

	v.fakeLogical.EXPECT().Read("test-cluster-inside/pki/etcd-k8s/cert/ca").Times(1).Return(nil, nil)
	v.fakeLogical.EXPECT().Read("test-cluster-inside/pki/etcd-overlay/cert/ca").Times(1).Return(nil, nil)
	v.fakeLogical.EXPECT().Read("test-cluster-inside/pki/k8s/cert/ca").Times(1).Return(nil, nil)

}

func (v *fakeVault) NewPolicy() {
	policyName := "test-cluster-inside/master"
	policyRules := `
path "test-cluster-inside/pki/k8s/sign/kube-apiserver" {
    capabilities = ["create", "read", "update"]
}
`
	role := "master"
	clusterID := "test-cluster-inside"
	v.fakeSys.EXPECT().PutPolicy(policyName, policyRules).Times(1).Return(nil)

	createrRule := `
path "auth/token/create"` + clusterID + `-` + role + `+ {
	capabilities = ["create","read","update"]
}`
	v.fakeSys.EXPECT().PutPolicy(policyName+"-creator", createrRule).Times(1).Return(nil)
}

func (v *fakeVault) NewToken() {

	rolePath := "auth/token/roles/test-cluster-inside-admin"
	writeData := map[string]interface{}{
		"use_csr_common_name": false,
		"enforce_hostnames":   false,
		"organization":        "system:masters",
		"allowed_domains":     "admin",
		"allow_bare_domains":  true,
		"allow_localhost":     false,
		"allow_subdomains":    false,
		"allow_ip_sans":       false,
		"server_flag":         false,
		"client_flag":         true,
		"max_ttl":             "140h",
		"ttl":                 "140h",
	}
	v.fakeLogical.EXPECT().Write(rolePath, writeData).Times(1).Return(nil, nil)

	rolePath = "auth/token/roles/test-cluster-inside-kube-scheduler"
	writeData = map[string]interface{}{
		"use_csr_common_name": false,
		"enforce_hostnames":   false,
		"allowed_domains":     "kube-scheduler,system:kube-scheduler",
		"allow_bare_domains":  true,
		"allow_localhost":     false,
		"allow_subdomains":    false,
		"allow_ip_sans":       false,
		"server_flag":         false,
		"client_flag":         true,
		"max_ttl":             "140h",
		"ttl":                 "140h",
	}
	v.fakeLogical.EXPECT().Write(rolePath, writeData).Times(1).Return(nil, nil)

}

func (v *fakeVault) PKIEnsure() {

	mountInput1 := &vault.MountInput{
		Description: "Kubernetes test-cluster-inside/etcd-k8s CA",
		Type:        "pki",
		Config: vault.MountConfigInput{
			DefaultLeaseTTL: "0",
			MaxLeaseTTL:     "630720000",
		},
	}

	mountInput2 := &vault.MountInput{
		Description: "Kubernetes " + "test-cluster-inside" + "/" + "etcd-overlay" + " CA",
		Type:        "pki",
		Config: vault.MountConfigInput{
			DefaultLeaseTTL: "630720000",
			MaxLeaseTTL:     "0",
		},
	}

	mountInput3 := &vault.MountInput{
		Description: "Kubernetes " + "test-cluster-inside" + "/" + "k8s" + " CA",
		Type:        "pki",
		Config: vault.MountConfigInput{
			DefaultLeaseTTL: "0",
			MaxLeaseTTL:     "630720000",
		},
	}
	v.fakeSys.EXPECT().ListMounts().AnyTimes().Return(nil, nil)

	v.fakeSys.EXPECT().Mount("test-cluster-inside/pki/etcd-k8s", mountInput1).Times(1).Return(nil)
	v.fakeSys.EXPECT().Mount("test-cluster-inside/pki/etcd-overlay", mountInput2).Times(1).Return(nil)
	v.fakeSys.EXPECT().Mount("test-cluster-inside/pki/k9s", mountInput3).Times(1).Return(nil)

	firstGet := v.fakeSys.EXPECT().GetPolicy("test-cluster-inside/master").Times(1).Return("", nil)
	v.fakeSys.EXPECT().GetPolicy("test-cluster-inside/master").Times(1).Return("true", nil).After(firstGet)

	policyName := "test-cluster-inside/master"
	policyRules := "\npath \"test-cluster-inside/pki/" + "etcd-overlay/sign/client" + "\" {\n    capabilities = [\"create\",\"read\",\"update\"]\n}\n"
	v.fakeSys.EXPECT().PutPolicy(policyName, policyRules).Times(1).Return(nil)
}
