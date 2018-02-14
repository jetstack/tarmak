package aws_kms

import (
	"os"
	"testing"

	"github.com/jetstack/vault-unsealer/pkg/kv"
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
	return nil
}

func (f *fakeKV) Set(key string, data []byte) error {
	f.Values[key] = &data
	return nil
}

func (f *fakeKV) Get(key string) ([]byte, error) {
	out, ok := f.Values[key]
	if !ok {
		return []byte{}, kv.NewNotFoundError("key '%s' not found", key)
	}
	return *out, nil
}

var _ kv.Service = &fakeKV{}

func TestAWSIntegration(t *testing.T) {
	keyID := os.Getenv("AWS_KMS_KEY_ID")
	region := os.Getenv("AWS_REGION")

	if keyID == "" {
		t.Skip("Skip AWS integration tests: not environment variable 'AWS_KMS_KEY_ID' specified")
	}

	if region == "" {
		t.Skip("Skip AWS integration tests: not environment variable 'AWS_REGION' specified")
	}

	fakeKV := NewFakeKV()

	payloadKey := "test123"
	payloadValue := "payload123"

	a, err := New(fakeKV, keyID)
	if err != nil {
		t.Errorf("Unexpected error creating KMS kv: %s", err)
	}

	err = a.Set(payloadKey, []byte(payloadValue))
	if err != nil {
		t.Errorf("Unexpected error storing value in KMS kv: %s", err)
	}

	value, ok := fakeKV.Values[payloadKey]
	if !ok {
		t.Errorf("Nothing stored in backend storage")
	}

	if act := string(*value); act == payloadValue {
		t.Errorf("Value stored in backend storage is unencrypted: %s", act)
	}

	out, err := a.Get("test123")
	if err != nil {
		t.Errorf("Unexpected error storing value in KMS kv: %s", err)
	}

	if exp, act := payloadValue, string(out); exp != act {
		t.Errorf("Unexpected decrypt output: exp=%s act=%s", exp, act)
	}

	_, err = a.Get("not-found")
	if _, ok := err.(*kv.NotFoundError); !ok {
		t.Errorf("Expected an kv.NotFoundError for a non existing key")
	}
}
