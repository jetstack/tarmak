// Copyright Jetstack Ltd. See LICENSE for details.
package hash

import (
	"bytes"
	"fmt"
	"io"
	"net/url"
	"path"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

const (
	S3HashObject   = "latest-puppet-hash"
	S3HashDir      = "puppet-manifests"
	S3LegacyObject = "puppet.tar.gz"
)

type Hash struct{}

func (h *Hash) GetManifest(hashPath string) (io.ReadCloser, error) {
	manifestURL, err := url.Parse(hashPath)
	if err != nil {
		return nil, err
	}

	// if we are pointing to the legacy object, change the key to point to the
	// hash object directory to get the latest hash
	bucket := manifestURL.Host
	key := manifestURL.Path
	if path.Base(manifestURL.Path) == S3LegacyObject {
		key = path.Join(path.Dir(key), S3HashDir)
	}
	key = path.Join(key, S3HashObject)

	cfg := aws.NewConfig()
	awsSession := session.New(cfg)
	s3Service := s3.New(awsSession)

	obj, err := s3Service.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("error getting s3 object '%s' in bucket '%s': %s", key, bucket, err)
	}

	b := new(bytes.Buffer)
	b.ReadFrom(obj.Body)

	key = fmt.Sprintf("%s/%s-puppet.tar.gz", path.Dir(key), b.String())
	obj, err = s3Service.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("error getting encrypted object by hash '%s' in bucket '%s': %s", key, bucket, err)
	}

	return obj.Body, nil
}

func (h *Hash) Name() string {
	return "hash"
}
