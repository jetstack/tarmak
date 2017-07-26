package vault

import (
	"fmt"
	"testing"

	"github.com/jetstack-experimental/vault-unsealer/pkg/kv"
)

type fakeKV struct {
	Values map[string]*[]byte
}

func NewFakeKV() *fakeKV {
	return &fakeKV{
		Values: map[string]*[]byte{},
	}
}

func (f *fakeKV) Test(key string) error {
	return fmt.Errorf("not-implemented")
}

func (f *fakeKV) Set(key string, data []byte) error {
	return fmt.Errorf("not-implemented")
}

func (f *fakeKV) Get(key string) ([]byte, error) {
	if key == "exists" {
		return []byte("data"), nil
	} else if key == "not-found" {
		return nil, kv.NewNotFoundError("not-found")

	}

	return nil, fmt.Errorf("not-implemented")
}

func TestKeyStoreNotFound(t *testing.T) {
	fakeKV := NewFakeKV()
	v := &vault{
		keyStore: fakeKV,
	}

	if !v.keyStoreNotFound("not-found") {
		t.Error("not returning true for notfound")
	}

	if v.keyStoreNotFound("exists") {
		t.Error("not returing false for existing")
	}

	if v.keyStoreNotFound("error") {
		t.Error("not returning false for error case")
	}
}
