// Copyright Jetstack Ltd. See LICENSE for details.
package hash

import (
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

const (
	S3HashObjest = "latest-puppet-hash"
)

type Hash struct{}

func (h *Hash) GetManifest(bucketPath string) (io.ReadCloser, error) {
	cfg := aws.NewConfig()
	awsSession := session.New(cfg)
	s3Service := s3.New(awsSession)

	obj, err := s3Service.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(bucketPath),
		Key:    aws.String(S3HashObjest),
	})
	if err != nil {
		return nil, fmt.Errorf("error getting s3 object '%s' in bucket '%s': %s", S3HashObjest, bucketPath, err)
	}

	b := make([]byte, 4096)
	n, err := obj.Body.Read(b)
	if err != nil {
		return nil, err
	}
	b = b[:n]

	objectPath := fmt.Sprintf("%s-puppet.tar.gz", b)
	obj, err = s3Service.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(bucketPath),
		Key:    aws.String(objectPath),
	})
	if err != nil {
		return nil, fmt.Errorf("error getting encrypted object by hash '%s' in bucket '%s': %s", objectPath, bucketPath, err)
	}

	return obj.Body, nil
}
