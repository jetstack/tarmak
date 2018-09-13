// Copyright Jetstack Ltd. See LICENSE for details.
package hash

import (
	"bytes"
	"fmt"
	"io"
	"net/url"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

const (
	S3HashObjest = "latest-puppet-hash"
)

type Hash struct{}

func (h *Hash) GetManifest(hashPath string) (io.ReadCloser, error) {
	manifestURL, err := url.Parse(hashPath)
	if err != nil {
		return nil, err
	}

	bucket := manifestURL.Host
	key := fmt.Sprintf("%s/%s", manifestURL.Path, S3HashObjest)

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

	key = fmt.Sprintf("%s/%s-puppet.tar.gz", manifestURL.Path, b.String())
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
