// Copyright Jetstack Ltd. See LICENSE for details.
package aws

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"

	tagControl "github.com/jetstack/tarmak/pkg/tagging_control"
)

const (
	identityEndpoint = "http://169.254.169.254/latest/dynamic/instance-identity"
	keyDir           = "/etc/ssh"
)

type AWSTags struct {
	log         *logrus.Entry
	environment string
}

func New(log *logrus.Entry, e string) *AWSTags {
	return &AWSTags{
		log:         log,
		environment: e,
	}
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

	request := &tagControl.TagInstanceRequest{
		KeySignatures:       sigs,
		PublicKeys:          pks,
		InstanceDocumentRaw: document,
		RSASigniture:        rsaSig,
	}

	if err := a.callLambdaFunction(request); err != nil {
		return err
	}

	a.log.Infof("successfully ensured instance tags")

	return nil
}

func (a *AWSTags) callLambdaFunction(request *tagControl.TagInstanceRequest) error {
	b, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %s", err)
	}

	document := new(ec2metadata.EC2InstanceIdentityDocument)
	err = json.Unmarshal(request.InstanceDocumentRaw, document)
	if err != nil {
		return fmt.Errorf("failed to unmarshal identity document: %s", err)
	}

	svc := lambda.New(session.New(&aws.Config{
		Region: aws.String(document.Region),
	}))

	resp, err := svc.Invoke(&lambda.InvokeInput{
		FunctionName: aws.String(fmt.Sprintf("%s_tarmak_tagging_control", a.environment)),
		Payload:      b,
	})
	if err != nil {
		return fmt.Errorf("failed to invoke lambda function: %s", err)
	}

	if resp.LogResult != nil {
		a.log.Debug(*resp.LogResult)
	}

	if resp.FunctionError != nil {
		return fmt.Errorf("tagging control function failed: %s: %s", *resp.FunctionError, resp.Payload)
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

		path := filepath.Join(keyDir, f.Name())

		fileData, err := ioutil.ReadFile(path)
		if err != nil {
			return nil, nil, err
		}

		// public key file
		if strings.HasSuffix(f.Name(), ".pub") {

			// ensure we do have a public key, not private
			// exit failure if not public, we have bad naming scheme so reject
			_, _, _, rest, err := ssh.ParseAuthorizedKey(fileData)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to parse local public key %s: %s", path, err)
			}

			if len(rest) != 0 {
				return nil, nil, fmt.Errorf("got rest parsing public key: %s", rest)
			}

			name := strings.TrimSuffix(f.Name(), filepath.Ext(f.Name()))

			a.log.Debugf("using public key %s", path)
			publicKeys[name] = fileData

			continue
		}

		// private key
		signer, err := ssh.ParsePrivateKey(fileData)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to parse local private key %s: %s", path, err)
		}

		sig, err := signer.Sign(rand.Reader, document)
		if err != nil {
			return nil, nil, err
		}

		a.log.Debugf("using signature from private key %s", path)
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
