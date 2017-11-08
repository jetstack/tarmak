package kubernetes

import (
	"fmt"
	"testing"

	"github.com/Sirupsen/logrus"
	"github.com/golang/mock/gomock"
	vault "github.com/hashicorp/vault/api"
)

type fakeVault struct {
	ctrl *gomock.Controller

	fakeVault   *MockVault
	fakeSys     *MockVaultSys
	fakeLogical *MockVaultLogical
	fakeAuth    *MockVaultAuth
	fakeToken   *MockVaultToken
}

func NewFakeVault(t *testing.T) *fakeVault {
	ctrl := gomock.NewController(t)

	v := &fakeVault{
		ctrl: ctrl,

		fakeVault:   NewMockVault(ctrl),
		fakeSys:     NewMockVaultSys(ctrl),
		fakeLogical: NewMockVaultLogical(ctrl),
		fakeAuth:    NewMockVaultAuth(ctrl),
		fakeToken:   NewMockVaultToken(ctrl),
	}

	v.fakeVault.EXPECT().Sys().AnyTimes().Return(v.fakeSys)
	v.fakeVault.EXPECT().Logical().AnyTimes().Return(v.fakeLogical)
	v.fakeVault.EXPECT().Auth().AnyTimes().Return(v.fakeAuth)
	v.fakeAuth.EXPECT().Token().AnyTimes().Return(v.fakeToken)

	return v
}

func (v *fakeVault) Kubernetes() *Kubernetes {
	k := New(nil, logrus.NewEntry(logrus.New()))
	k.SetClusterID("test-cluster-inside")
	k.vaultClient = v.fakeVault
	return k
}

func (v *fakeVault) Finish() {
	v.ctrl.Finish()
}

func (v *fakeVault) Ensure() {
	v.fakeSys.EXPECT().ListMounts().AnyTimes().Return(nil, nil)

	v.fakeSys.EXPECT().Mount("test-cluster-inside/pki/etcd-k8s", gomock.Any()).Times(1).Return(nil)
	v.fakeSys.EXPECT().Mount("test-cluster-inside/pki/etcd-overlay", gomock.Any()).Times(1).Return(nil)
	v.fakeSys.EXPECT().Mount("test-cluster-inside/pki/k8s", gomock.Any()).Times(1).Return(nil)
	v.fakeSys.EXPECT().Mount("test-cluster-inside/secrets", gomock.Any()).Times(1).Return(nil)

	v.fakeLogical.EXPECT().Read("test-cluster-inside/pki/etcd-k8s/cert/ca").Times(1).Return(nil, nil)
	v.fakeLogical.EXPECT().Write("test-cluster-inside/pki/etcd-k8s/root/generate/internal", gomock.Any()).Times(1).Return(nil, nil)

	v.fakeLogical.EXPECT().Read("test-cluster-inside/pki/etcd-overlay/cert/ca").Times(1).Return(nil, nil)
	v.fakeLogical.EXPECT().Write("test-cluster-inside/pki/etcd-overlay/root/generate/internal", gomock.Any()).Times(1).Return(nil, nil)

	v.fakeLogical.EXPECT().Read("test-cluster-inside/pki/k8s/cert/ca").Times(1).Return(nil, nil)
	v.fakeLogical.EXPECT().Write("test-cluster-inside/pki/k8s/root/generate/internal", gomock.Any()).Times(1).Return(nil, nil)

	v.fakeLogical.EXPECT().Read("test-cluster-inside/secrets/service-accounts").Times(1).Return(nil, nil)
	v.fakeLogical.EXPECT().Write("test-cluster-inside/secrets/service-accounts", gomock.Any()).Times(1).Return(nil, nil)

	v.fakeLogical.EXPECT().Write("test-cluster-inside/pki/etcd-k8s/roles/client", gomock.Any()).Times(1).Return(nil, nil)
	v.fakeLogical.EXPECT().Write("test-cluster-inside/pki/etcd-k8s/roles/server", gomock.Any()).Times(1).Return(nil, nil)

	v.fakeLogical.EXPECT().Write("test-cluster-inside/pki/etcd-overlay/roles/client", gomock.Any()).Times(1).Return(nil, nil)
	v.fakeLogical.EXPECT().Write("test-cluster-inside/pki/etcd-overlay/roles/server", gomock.Any()).Times(1).Return(nil, nil)

	v.fakeLogical.EXPECT().Write("test-cluster-inside/pki/k8s/roles/admin", gomock.Any()).Times(1).Return(nil, nil)
	v.fakeLogical.EXPECT().Write("test-cluster-inside/pki/k8s/roles/kube-apiserver", gomock.Any()).Times(1).Return(nil, nil)
	v.fakeLogical.EXPECT().Write("test-cluster-inside/pki/k8s/roles/kube-scheduler", gomock.Any()).Times(1).Return(nil, nil)
	v.fakeLogical.EXPECT().Write("test-cluster-inside/pki/k8s/roles/kube-controller-manager", gomock.Any()).Times(1).Return(nil, nil)
	v.fakeLogical.EXPECT().Write("test-cluster-inside/pki/k8s/roles/kube-proxy", gomock.Any()).Times(1).Return(nil, nil)
	last := v.fakeLogical.EXPECT().Write("test-cluster-inside/pki/k8s/roles/kubelet", gomock.Any()).Times(1).Return(nil, nil)

	v.fakeSys.EXPECT().PutPolicy(gomock.Any(), gomock.Any()).AnyTimes().Return(nil).After(last)

	v.fakeLogical.EXPECT().Read(gomock.Any()).AnyTimes().Return(nil, nil).After(last)
	v.fakeLogical.EXPECT().Write(gomock.Any(), gomock.Any()).AnyTimes().Return(nil, nil).After(last)
	v.fakeToken.EXPECT().CreateOrphan(gomock.Any()).AnyTimes().Return(&vault.Secret{
		Auth: &vault.SecretAuth{
			ClientToken: "my-new-token",
		},
	}, nil).After(last)
}

func (v *fakeVault) NewToken() {
	writeData := map[string]interface{}{"use_csr_common_name": false, "enforce_hostnames": false,
		"organization":       "system:masters",
		"allowed_domains":    "admin",
		"allow_bare_domains": true,
		"allow_localhost":    false,
		"allow_subdomains":   false,
		"allow_ip_sans":      false,
		"server_flag":        false,
		"client_flag":        true,
		"max_ttl":            "31536000s",
		"ttl":                "31536000s",
	}

	writeData2 := map[string]interface{}{
		"allow_localhost": false,
		"server_flag":     false,
		"max_ttl":         "2592000s",
		"ttl":             "2592000s",
		"use_csr_common_name": false,
		"allow_bare_domains":  true,
		"allow_subdomains":    false,
		"allow_ip_sans":       true,
		"allowed_domains":     "kube-scheduler,system:kube-scheduler",
		"enforce_hostnames":   false,
		"client_flag":         true,
	}

	v.fakeLogical.EXPECT().Write("test-cluster-inside/pki/k8s/roles/admin", writeData).AnyTimes().Return(nil, nil)
	v.fakeLogical.EXPECT().Write("test-cluster-inside/pki/k8s/roles/kube-scheduler", writeData2).AnyTimes().Return(nil, nil)
}

func (v *fakeVault) PKIEnsure() {
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

	v.fakeSys.EXPECT().Mount("test-cluster-inside/pki/etcd-k8s", mountInput1).Times(1).Return(nil)
	v.fakeSys.EXPECT().Mount("test-cluster-inside/pki/etcd-overlay", mountInput2).Times(1).Return(nil)
	v.fakeSys.EXPECT().Mount("test-cluster-inside/pki/k8s", mountInput3).Times(1).Return(nil)

	v.fakeLogical.EXPECT().Read(gomock.Any()).AnyTimes().Return(nil, nil)

	v.fakeLogical.EXPECT().Write(gomock.Any(), gomock.Any()).AnyTimes().Return(nil, nil)

	v.fakeSys.EXPECT().Mount("test-cluster-inside/secrets", gomock.Any()).Times(1).Return(nil)

	v.fakeSys.EXPECT().PutPolicy(gomock.Any(), gomock.Any()).AnyTimes().Return(nil)
	v.fakeToken.EXPECT().CreateOrphan(gomock.Any()).AnyTimes().Return(&vault.Secret{
		Auth: &vault.SecretAuth{
			ClientToken: "my-new-token",
		},
	}, nil)

	v.fakeSys.EXPECT().GetPolicy(gomock.Any()).Return("", nil)
	v.fakeSys.EXPECT().GetPolicy(gomock.Any()).Return("policy", nil)

	first := v.fakeSys.EXPECT().Mount("test-cluster-inside/pki/wrong-type-pki", gomock.Any()).Times(1).Return(nil)
	v.fakeSys.EXPECT().Mount("test-cluster-inside/pki/wrong-type-pki", gomock.Any()).Times(1).Return(fmt.Errorf("wrong type")).After(first)
}
