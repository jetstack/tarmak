package aws_ssm

import (
	"os"
	"testing"

	"github.com/jetstack-experimental/vault-unsealer/pkg/kv"
)

func TestAWSIntegration(t *testing.T) {
	region := os.Getenv("AWS_REGION")

	if region == "" {
		t.Skip("Skip AWS integration tests: not environment variable 'AWS_REGION' specified")
	}

	payloadKey := "test123"
	payloadValue := "payload123"

	a, err := New("test-integration-")
	if err != nil {
		t.Errorf("Unexpected error creating SSM kv: %s", err)
	}

	// graceful set (in case it's already existing)
	err = a.Set(payloadKey, []byte(payloadValue))
	if err != nil {
		t.Errorf("Unexpected error storing value in SSM kv: %s", err)
	}

	// this should also work and overwrite a key
	err = a.Set(payloadKey, []byte(payloadValue))
	if err != nil {
		t.Errorf("Unexpected error storing value in SSM kv: %s", err)
	}

	// this deletes the key
	err = a.Delete(payloadKey)
	if err != nil {
		t.Errorf("Unexpected error storing value in SSM kv: %s", err)
	}

	_, err = a.Get(payloadKey)
	if _, ok := err.(*kv.NotFoundError); !ok {
		t.Errorf("Expected an kv.NotFoundError for a non existing key")
	}

	err = a.Set(payloadKey, []byte(payloadValue))
	if err != nil {
		t.Errorf("Unexpected error storing value in SSM kv: %s", err)
	}

	out, err := a.Get("test123")
	if err != nil {
		t.Errorf("Unexpected error storing value in SSM kv: %s", err)
	}

	if exp, act := payloadValue, string(out); exp != act {
		t.Errorf("Unexpected decrypt output: exp=%s act=%s", exp, act)
	}

}
