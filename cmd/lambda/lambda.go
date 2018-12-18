package main

import (
	"context"
	//"crypto"
	//"crypto/ecdsa"
	//"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"math/big"
	//"encoding/json"
	"encoding/pem"
	//"errors"
	"fmt"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"golang.org/x/crypto/ssh"
	//"golang.org/x/crypto/ed25519"
	//"github.com/fullsailor/pkcs7"
)

const (
	tagSize = 256
)

var (
	AWSCert = []byte(`-----BEGIN CERTIFICATE-----
MIIDIjCCAougAwIBAgIJAKnL4UEDMN/FMA0GCSqGSIb3DQEBBQUAMGoxCzAJBgNV
BAYTAlVTMRMwEQYDVQQIEwpXYXNoaW5ndG9uMRAwDgYDVQQHEwdTZWF0dGxlMRgw
FgYDVQQKEw9BbWF6b24uY29tIEluYy4xGjAYBgNVBAMTEWVjMi5hbWF6b25hd3Mu
Y29tMB4XDTE0MDYwNTE0MjgwMloXDTI0MDYwNTE0MjgwMlowajELMAkGA1UEBhMC
VVMxEzARBgNVBAgTCldhc2hpbmd0b24xEDAOBgNVBAcTB1NlYXR0bGUxGDAWBgNV
BAoTD0FtYXpvbi5jb20gSW5jLjEaMBgGA1UEAxMRZWMyLmFtYXpvbmF3cy5jb20w
gZ8wDQYJKoZIhvcNAQEBBQADgY0AMIGJAoGBAIe9GN//SRK2knbjySG0ho3yqQM3
e2TDhWO8D2e8+XZqck754gFSo99AbT2RmXClambI7xsYHZFapbELC4H91ycihvrD
jbST1ZjkLQgga0NE1q43eS68ZeTDccScXQSNivSlzJZS8HJZjgqzBlXjZftjtdJL
XeE4hwvo0sD4f3j9AgMBAAGjgc8wgcwwHQYDVR0OBBYEFCXWzAgVyrbwnFncFFIs
77VBdlE4MIGcBgNVHSMEgZQwgZGAFCXWzAgVyrbwnFncFFIs77VBdlE4oW6kbDBq
MQswCQYDVQQGEwJVUzETMBEGA1UECBMKV2FzaGluZ3RvbjEQMA4GA1UEBxMHU2Vh
dHRsZTEYMBYGA1UEChMPQW1hem9uLmNvbSBJbmMuMRowGAYDVQQDExFlYzIuYW1h
em9uYXdzLmNvbYIJAKnL4UEDMN/FMAwGA1UdEwQFMAMBAf8wDQYJKoZIhvcNAQEF
BQADgYEAFYcz1OgEhQBXIwIdsgCOS8vEtiJYF+j9uO6jz7VOmJqO+pRlAbRlvY8T
C1haGgSI/A1uZUKs/Zfnph0oEI0/hu1IIJ/SKBDtN5lvmZ/IzbOPIJWirlsllQIQ
7zvWbGd9c9+Rm3p04oTvhup99la7kZqevJK0QRdD/6NpCKsqP/0=
-----END CERTIFICATE-----`)
)

type EC2InstanceIdentityDocument struct {
	DevpayProductCodes []string  `json:"devpayProductCodes"`
	AvailabilityZone   string    `json:"availabilityZone"`
	PrivateIP          string    `json:"privateIp"`
	Version            string    `json:"version"`
	Region             string    `json:"region"`
	InstanceID         string    `json:"instanceId"`
	BillingProducts    []string  `json:"billingProducts"`
	InstanceType       string    `json:"instanceType"`
	AccountID          string    `json:"accountId"`
	PendingTime        time.Time `json:"pendingTime"`
	ImageID            string    `json:"imageId"`
	KernelID           string    `json:"kernelId"`
	RamdiskID          string    `json:"ramdiskId"`
	Architecture       string    `json:"architecture"`
}

type TagInstanceRequest struct {
	PublicKeys          map[string][]byte         `json:"publicKeys"`
	KeySignatures       map[string]*ssh.Signature `json:"KeySignatures"`
	InstanceDocumentRaw []byte                    `json:"document"`
	RSASigniture        []byte                    `json:"rsaSigniture"`
}

type ECDSASignature struct {
	R *big.Int `json:"R"`
	S *big.Int `json:"S"`
}

func HandleRequest(ctx context.Context, t *TagInstanceRequest) error {
	if err := t.verify(); err != nil {
		return err
	}

	tags := t.createTags()
	exists, err := t.checkTagsAgainstInstance(tags)
	if err != nil || exists {
		return err
	}

	// attach tags to ec2 instance using real call
	//err := ec2.Tag{
	//	InstanceID: t.InstanceDocument.InstanceID,
	//	Tags: ....
	//}
	// if err != nil {
	//	return err
	//}
	return nil
}

// verify the rsa signature against the instance identity content and AWS global
// cert
func (t *TagInstanceRequest) verify() error {
	block, rest := pem.Decode([]byte(AWSCert))
	if len(rest) != 0 {
		return fmt.Errorf("expected to fully parse AWS certificate but had remainder: %s", rest)
	}

	awsCaCert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return err
	}

	rsaSignature, err := base64.StdEncoding.DecodeString(string(t.RSASigniture))
	if err != nil {
		return err
	}

	err = awsCaCert.CheckSignature(x509.SHA256WithRSA, t.InstanceDocumentRaw, rsaSignature)
	if err != nil {
		return err
	}

	for _, v := range t.PublicKeys {
		_, _, _, rest, err := ssh.ParseAuthorizedKey(v)
		if err != nil {
			return fmt.Errorf("failed to parse public key: %s", err)
		}

		if len(rest) != 0 {
			return fmt.Errorf("got rest parsing public key: %s", rest)
		}
	}

	return nil
}

// check generated tags against the ec2 instance
// if existing and match exit gracefully
// if miss match, exit failure
// if not exist, we need to create
func (t *TagInstanceRequest) checkTagsAgainstInstance(tags map[string][]byte) (tagsExist bool, err error) {
	return false, nil
}

// split up public keys into correct sizes for AWS tags
func (t *TagInstanceRequest) createTags() map[string][]byte {
	tags := make(map[string][]byte)
	for keyName, data := range t.PublicKeys {
		data = append(data, []byte("==EOF")...)
		for i := 0; i < len(data); i += tagSize {
			end := i + tagSize
			if end > len(data) {
				end = len(data)
			}
			tagName := fmt.Sprintf("PublicKey_%s_%d", keyName, i/tagSize)
			tags[tagName] = data[i:end]
		}
	}
	return tags
}
func main() {
	lambda.Start(HandleRequest)
}
