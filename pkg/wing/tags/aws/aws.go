package aws

import (
	"crypto/rand"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"path/filepath"
	"strings"

	"github.com/jetstack/tarmak/cmd/tagging_control/cmd"
	"golang.org/x/crypto/ssh"
)

const (
	identityEndpoint = "http://169.254.169.254/latest/dynamic/instance-identity"
	keyDir           = "/etc/ssh"
)

func EnsureMachineTags() error {
	document, err := requestData("document")
	if err != nil {
		return err
	}

	rsaSig, err := requestData("signature")
	if err != nil {
		return err
	}

	pks, sigs, err := fetchLocalKeys(document)
	if err != nil {
		return err
	}

	request := &cmd.TagInstanceRequest{
		KeySignatures:       sigs,
		PublicKeys:          pks,
		InstanceDocumentRaw: document,
		RSASigniture:        rsaSig,
	}

	if err := callLambdaFunction(request); err != nil {
		return err
	}

	return nil
}

func callLambdaFunction(request *cmd.TagInstanceRequest) error {
	return nil
}

func fetchLocalKeys(document []byte) (map[string][]byte, map[string]*ssh.Signature, error) {
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

		// public key file
		if strings.HasSuffix(f.Name(), ".pub") {
			name := strings.TrimSuffix(f.Name(), filepath.Ext(f.Name()))
			publicKeys[name] = fileData

			continue
		}

		// private key
		block, rest := pem.Decode(fileData)
		if len(rest) != 0 {
			return nil, nil, fmt.Errorf("expected to fully parse local private key but had remainder %s: %s",
				f.Name(), rest)
		}

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

func requestData(file string) ([]byte, error) {
	u, err := url.Parse(identityEndpoint)
	if err != nil {
		return nil, err
	}
	u.Path = path.Join(u.Path, file)

	res, err := http.Get(u.Path)
	if err != nil {
		return nil, err
	}

	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	return data, nil
}
