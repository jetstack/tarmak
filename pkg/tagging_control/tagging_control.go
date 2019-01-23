// Copyright Jetstack Ltd. See LICENSE for details.
package tagging_control

import (
	"context"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"golang.org/x/crypto/ssh"
)

const (
	tagSize   = 255
	tagPrefix = "tarmak.io"

	AWSCACert = `-----BEGIN CERTIFICATE-----
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
-----END CERTIFICATE-----`
)

type TagInstanceRequest struct {
	PublicKeys    map[string][]byte         `json:"publicKeys"`
	KeySignatures map[string]*ssh.Signature `json:"keySignatures"`

	InstanceDocumentRaw []byte `json:"document"`
	RSASignature        []byte `json:"rsaSignature"`
}

type Handler struct {
	request  *TagInstanceRequest
	ec2      *ec2.EC2
	document *ec2metadata.EC2InstanceIdentityDocument
}

func HandleRequests(ctx context.Context, request TagInstanceRequest) error {
	h := &Handler{
		request: &request,
	}

	if err := h.verify(); err != nil {
		// until we verify the instance, we won't reply any meaningful error message
		fmt.Printf("failed to verify instance: %s", err)
		return errors.New("rejected")
	}

	document := new(ec2metadata.EC2InstanceIdentityDocument)
	err := json.Unmarshal(h.request.InstanceDocumentRaw, document)
	if err != nil {
		return err
	}
	h.document = document

	tags := h.createTags()
	exists, err := h.checkTagsAgainstInstance(tags)
	if err != nil || exists {
		return err
	}

	// tags do not exist on instance so create
	svc, err := h.EC2()
	if err != nil {
		return err
	}

	var ec2Tags []*ec2.Tag
	for k, v := range tags {
		ec2Tags = append(ec2Tags, &ec2.Tag{
			Key:   aws.String(k),
			Value: aws.String(v),
		})
	}

	_, err = svc.CreateTags(&ec2.CreateTagsInput{
		Tags: ec2Tags,
		Resources: []*string{
			aws.String(h.document.InstanceID),
		},
	})
	if err != nil {
		return fmt.Errorf("failed to create tags for instance %s: %s", h.document.InstanceID, err)
	}

	return nil
}

// verify the rsa signature against the instance identity content and AWS global
// cert
func (h *Handler) verify() error {
	block, rest := pem.Decode([]byte(AWSCACert))
	if len(rest) != 0 {
		return fmt.Errorf("expected to fully parse AWS certificate but had remainder: %s", rest)
	}

	awsCaCert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return err
	}

	rsaSignature, err := base64.StdEncoding.DecodeString(string(h.request.RSASignature))
	if err != nil {
		return err
	}

	err = awsCaCert.CheckSignature(x509.SHA256WithRSA, h.request.InstanceDocumentRaw, rsaSignature)
	if err != nil {
		return fmt.Errorf("failed to verify identity document signature against AWS: %s", err)
	}

	// verify ssh keys
	for k, v := range h.request.PublicKeys {
		pk, _, _, rest, err := ssh.ParseAuthorizedKey(v)
		if err != nil {
			return fmt.Errorf("failed to parse public key: %s", err)
		}

		if len(rest) != 0 {
			return fmt.Errorf("got rest parsing public key: %s", rest)
		}

		sig, ok := h.request.KeySignatures[k]
		if !ok {
			return fmt.Errorf("did not receive signature for public key %s", k)
		}

		err = pk.Verify(h.request.InstanceDocumentRaw, sig)
		if err != nil {
			return fmt.Errorf("could not verify public key %s: %s", k, err)
		}
	}

	return nil
}

// check generated tags against the ec2 instance
// if existing and match exit gracefully
// if miss match, exit failure
// if not exist, we need to create
func (h *Handler) checkTagsAgainstInstance(tags map[string]string) (tagsExist bool, err error) {
	svc, err := h.EC2()
	if err != nil {
		return false, err
	}

	out, err := svc.DescribeTags(&ec2.DescribeTagsInput{
		Filters: []*ec2.Filter{
			&ec2.Filter{
				Name:   aws.String("resource-id"),
				Values: []*string{aws.String(h.document.InstanceID)},
			},
			&ec2.Filter{
				Name:   aws.String("resource-type"),
				Values: []*string{aws.String("instance")},
			},
		},
	})
	if err != nil {
		return false, fmt.Errorf("failed to list tags of instance %s: %s", h.document.InstanceID, err)
	}

	for _, existingTag := range out.Tags {

		// check for public key tag prefix
		if strings.HasPrefix(*existingTag.Key, tagPrefix) {

			// tags exist on the instance with the matching prefix
			tagsExist = true

			value, ok := tags[*existingTag.Key]

			// given tags do not include an existing tag
			if !ok {
				return true, errors.New("mismatch tags, rejected")
			}

			// existing and given tag mismatch values
			if *existingTag.Value != value {
				return true, errors.New("mismatch tags, rejected")
			}
		}
	}

	return tagsExist, nil
}

// split up public keys into correct sizes for AWS tags
func (h *Handler) createTags() map[string]string {
	tags := make(map[string]string)

	for keyName, data := range h.request.PublicKeys {
		data = append(data, []byte("==EOF")...)

		for i := 0; i < len(data); i += tagSize {
			end := i + tagSize

			if end > len(data) {
				end = len(data)
			}

			tagName := fmt.Sprintf("%s/%s-%d", tagPrefix, keyName, i/tagSize)
			tags[tagName] = string(data[i:end])
		}
	}

	return tags
}

func (h *Handler) EC2() (*ec2.EC2, error) {
	if h.ec2 != nil {
		return h.ec2, nil
	}

	region := os.Getenv("AWS_REGION")
	session, err := session.NewSession(&aws.Config{
		Region: &region,
	})
	if err != nil {
		return nil, err
	}

	h.ec2 = ec2.New(session)

	return h.ec2, nil
}
