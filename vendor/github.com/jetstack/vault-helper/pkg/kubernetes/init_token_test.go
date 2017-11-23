package kubernetes

import (
	"testing"

	"github.com/golang/mock/gomock"
	vault "github.com/hashicorp/vault/api"
)

type tokenCreateRequestMatcher struct {
	ID   string
	name string
}

func (tcrm *tokenCreateRequestMatcher) String() string {
	return "matcher"
}

func (tcrm *tokenCreateRequestMatcher) Matches(x interface{}) bool {
	tcr, ok := x.(*vault.TokenCreateRequest)
	if !ok {
		return false
	}

	if tcrm.ID != tcr.ID {
		return false
	}

	return true
}

// tests a not yet existing init token, with random generated token
func TestInitToken_Ensure_NoExpectedToken_NotExisting(t *testing.T) {
	fv := NewFakeVault(t)
	defer fv.Finish()

	i := &InitToken{
		Role:          "etcd",
		Policies:      []string{"etcd"},
		kubernetes:    fv.Kubernetes(),
		ExpectedToken: "",
	}

	// expects a read and vault says secret is not existing
	genericPath := "test-cluster-inside/secrets/init_token_etcd"
	fv.fakeLogical.EXPECT().Read(genericPath).Return(
		nil,
		nil,
	)

	// expect a create new orphan
	fv.fakeToken.EXPECT().CreateOrphan(&tokenCreateRequestMatcher{}).Return(&vault.Secret{
		Auth: &vault.SecretAuth{
			ClientToken: "my-new-random-token",
		},
	}, nil)

	// expect a write of the new token
	fv.fakeLogical.EXPECT().Write(genericPath, map[string]interface{}{"init_token": "my-new-random-token"}).Return(
		nil,
		nil,
	)

	InitTokenEnsure_EXPECTs(fv)

	err := i.Ensure()
	if err != nil {
		t.Error("unexpected error: ", err)
	}

	token, err := i.InitToken()
	if err != nil {
		t.Error("unexpected error: ", err)
	}

	if exp, act := "my-new-random-token", token; exp != act {
		t.Errorf("unexpected token: act=%s exp=%s", act, exp)
	}

	return
}

// expected token not set, init token already exists
func TestInitToken_Ensure_NoExpectedToken_AlreadyExisting(t *testing.T) {
	fv := NewFakeVault(t)
	defer fv.Finish()

	i := &InitToken{
		Role:          "etcd",
		Policies:      []string{"etcd"},
		kubernetes:    fv.Kubernetes(),
		ExpectedToken: "",
	}

	// expect a read and vault says secret is existing
	genericPath := "test-cluster-inside/secrets/init_token_etcd"
	fv.fakeLogical.EXPECT().Read(genericPath).Return(
		&vault.Secret{
			Data: map[string]interface{}{"init_token": "existing-token"},
		},
		nil,
	)

	InitTokenEnsure_EXPECTs(fv)

	err := i.Ensure()
	if err != nil {
		t.Error("unexpected error: ", err)
	}

	token, err := i.InitToken()
	if err != nil {
		t.Error("unexpected error: ", err)
	}

	if exp, act := "existing-token", token; exp != act {
		t.Errorf("unexpected token: act=%s exp=%s", act, exp)
	}

	return
}

// excpected token set, init token already exists and it's matching
func TestInitToken_Ensure_ExpectedToken_Existing_Match(t *testing.T) {
	fv := NewFakeVault(t)
	defer fv.Finish()

	i := &InitToken{
		Role:          "etcd",
		Policies:      []string{"etcd"},
		kubernetes:    fv.Kubernetes(),
		ExpectedToken: "expected-token",
	}

	// expect a read and vault says secret is existing
	genericPath := "test-cluster-inside/secrets/init_token_etcd"
	fv.fakeLogical.EXPECT().Read(genericPath).Return(
		&vault.Secret{
			Data: map[string]interface{}{"init_token": "expected-token"},
		},
		nil,
	)

	InitTokenEnsure_EXPECTs(fv)

	err := i.Ensure()
	if err != nil {
		t.Error("unexpected error: ", err)
	}

	token, err := i.InitToken()
	if err != nil {
		t.Error("unexpected error: ", err)
	}

	if exp, act := "expected-token", token; exp != act {
		t.Errorf("unexpected token: act=%s exp=%s", act, exp)
	}

	return
}

// expected token set, init token doesn't exist
func TestInitToken_Ensure_ExpectedToken_NotExisting(t *testing.T) {
	fv := NewFakeVault(t)
	defer fv.Finish()

	i := &InitToken{
		Role:          "etcd",
		Policies:      []string{"etcd"},
		kubernetes:    fv.Kubernetes(),
		ExpectedToken: "expected-token",
	}

	// expect a new token creation
	fv.fakeToken.EXPECT().CreateOrphan(&tokenCreateRequestMatcher{ID: "expected-token"}).Return(&vault.Secret{
		Auth: &vault.SecretAuth{
			ClientToken: "expected-token",
		},
	}, nil)

	// expect a read and vault says secret is not existing, then after it is written to return token
	genericPath := "test-cluster-inside/secrets/init_token_etcd"
	gomock.InOrder(
		fv.fakeLogical.EXPECT().Read(genericPath).Return(
			nil,
			nil,
		).MinTimes(1),
		// expect a write of the new token from user flag
		fv.fakeLogical.EXPECT().Write(genericPath, map[string]interface{}{"init_token": "expected-token"}).Return(
			nil,
			nil,
		),
		// allow read out of token from user
		fv.fakeLogical.EXPECT().Read(genericPath).AnyTimes().Return(
			&vault.Secret{
				Data: map[string]interface{}{"init_token": "expected-token"},
			},
			nil,
		),
	)

	InitTokenEnsure_EXPECTs(fv)

	err := i.Ensure()
	if err != nil {
		t.Error("unexpected error: ", err)
	}

	token, err := i.InitToken()
	if err != nil {
		t.Error("unexpected error: ", err)
	}

	if exp, act := "expected-token", token; exp != act {
		t.Errorf("unexpected token: act=%s exp=%s", act, exp)
	}

	return
}

// General policy and write calls when init token ensuring
func InitTokenEnsure_EXPECTs(fv *fakeVault) {
	fv.fakeLogical.EXPECT().Write("auth/token/roles/test-cluster-inside-etcd", gomock.Any()).AnyTimes().Return(nil, nil)
	fv.fakeSys.EXPECT().PutPolicy(gomock.Any(), gomock.Any()).AnyTimes().Return(nil)
}
