// Copyright Jetstack Ltd. See LICENSE for details.
package s3

import (
	"fmt"
	"io"
	"net/url"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

type S3 struct{}

func (s *S3) GetManifest(manifestString string) (io.ReadCloser, error) {
	manifestURL, err := url.Parse(manifestString)
	if err != nil {
		return nil, err
	}

	bucket := manifestURL.Host
	key := manifestURL.Path

	cfg := aws.NewConfig()
	awsSession := session.New(cfg)
	s3Service := s3.New(awsSession)

	result, err := s3Service.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("error getting s3 object '%s' in bucket '%s': %s", key, bucket, err)
	}

	return result.Body, nil
}

func (s *S3) Name() string {
	return "s3"
}
