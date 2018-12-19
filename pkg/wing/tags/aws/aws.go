// Copyright Jetstack Ltd. See LICENSE for details.
package aws

import (
	"crypto/rand"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/lambda"
	"golang.org/x/crypto/ssh"

	"github.com/jetstack/tarmak/cmd/tagging_control/cmd"
)

const (
	identityEndpoint = "http://169.254.169.254/latest/dynamic/instance-identity"
	keyDir           = "/etc/ssh"
)

type AWSTags struct{}

func New() *AWSTags {
	return new(AWSTags)
}

func (a *AWSTags) EnsureMachineTags() error {
	document, err := a.requestData("document")
	if err != nil {
		return err
	}

	rsaSig, err := a.requestData("signature")
	if err != nil {
		return err
	}

	pks, sigs, err := a.fetchLocalKeys(document)
	if err != nil {
		return err
	}

	request := &cmd.TagInstanceRequest{
		KeySignatures:       sigs,
		PublicKeys:          pks,
		InstanceDocumentRaw: document,
		RSASigniture:        rsaSig,
	}

	if err := a.callLambdaFunction(request); err != nil {
		return err
	}

	return nil
}

func (a *AWSTags) callLambdaFunction(request *cmd.TagInstanceRequest) error {
	b, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %s", err)
	}

	svc := lambda.New(session.New(aws.NewConfig()))
	_, err = svc.Invoke(&lambda.InvokeInput{
		FunctionName: aws.String("tagging_control"),
		Payload:      b,
	})

	if err != nil {
		return fmt.Errorf("failed to invoke lambda function: %s", err)
	}

	return nil
}

func (a *AWSTags) fetchLocalKeys(document []byte) (map[string][]byte, map[string]*ssh.Signature, error) {
	fs, err := ioutil.ReadDir(keyDir)
	if err != nil {
		return nil, nil, err
	}

	publicKeys := make(map[string][]byte)
	sigs := make(map[string]*ssh.Signature)
	for _, f := range fs {

		// not a key file
		if f.IsDir() || !strings.HasPrefix(f.Name(), "ssh_host") {
			continue
		}

		fileData, err := ioutil.ReadFile(f.Name())
		if err != nil {
			return nil, nil, err
		}

		block, rest := pem.Decode(fileData)
		if len(rest) != 0 {
			return nil, nil, fmt.Errorf("expected to fully parse local key file but had remainder %s: %s",
				f.Name(), rest)
		}

		// public key file
		if strings.HasSuffix(f.Name(), ".pub") {
			// ensure we do have a public key, not private
			_, err := ssh.ParsePublicKey(block.Bytes)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to parse local public key %s: %s", f.Name(), err)
			}

			name := strings.TrimSuffix(f.Name(), filepath.Ext(f.Name()))
			publicKeys[name] = fileData

			continue
		}

		// private key
		signer, err := ssh.ParsePrivateKey(block.Bytes)
		if err != nil {
			return nil, nil, err
		}

		sig, err := signer.Sign(rand.Reader, document)
		if err != nil {
			return nil, nil, err
		}

		sigs[f.Name()] = sig
	}

	return publicKeys, sigs, nil
}

func (a *AWSTags) requestData(file string) ([]byte, error) {
	res, err := http.Get(fmt.Sprintf("%s/%s", identityEndpoint, file))
	if err != nil {
		return nil, err
	}

	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	return data, nil
}
