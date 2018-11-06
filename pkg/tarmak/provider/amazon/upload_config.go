// Copyright Jetstack Ltd. See LICENSE for details.
package amazon

import (
	"bytes"
	"fmt"
	"io"
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/aws/aws-sdk-go/service/s3"

	"github.com/jetstack/tarmak/pkg/tarmak/interfaces"
)

// This uploads the main configuration to the S3 bucket
func (a *Amazon) UploadConfiguration(cluster interfaces.Cluster, stateFile io.ReadSeeker, md5Hash string) error {
	svcKMS, err := a.KMS()
	if err != nil {
		return err
	}

	k, err := svcKMS.DescribeKey(&kms.DescribeKeyInput{
		KeyId: aws.String(a.SecretsKMSName()),
	})
	if err != nil {
		return fmt.Errorf("error looking for tarmak secrets kms alias '%s': %s", a.SecretsKMSName(), err)
	}

	bucketName := fmt.Sprintf(
		"%s%s-%s-secrets",
		a.conf.Amazon.BucketPrefix,
		cluster.Environment().Name(),
		a.Region(),
	)

	svc, err := a.S3()
	if err != nil {
		return err
	}

	manifestKey := filepath.Join(cluster.ClusterName(), "puppet.tar.gz")
	_, err = svc.PutObject(&s3.PutObjectInput{
		Bucket:               aws.String(bucketName),
		Key:                  aws.String(manifestKey),
		Body:                 stateFile,
		ServerSideEncryption: aws.String(s3.ServerSideEncryptionAwsKms),
		SSEKMSKeyId:          aws.String(*k.KeyMetadata.Arn),
	})
	if err != nil {
		return err
	}

	if _, err := stateFile.Seek(0, 0); err != nil {
		return fmt.Errorf("failed to rewind puppet state file: %s", err)
	}

	dirPath := filepath.Join(cluster.ClusterName(), "puppet-manifests")
	hashPointerKey := filepath.Join(dirPath, "latest-puppet-hash")
	manifestKey = filepath.Join(dirPath, fmt.Sprintf("%s-puppet.tar.gz", md5Hash))
	_, err = svc.PutObject(&s3.PutObjectInput{
		Bucket:               aws.String(bucketName),
		Key:                  aws.String(manifestKey),
		Body:                 stateFile,
		ServerSideEncryption: aws.String(s3.ServerSideEncryptionAwsKms),
		SSEKMSKeyId:          aws.String(*k.KeyMetadata.Arn),
	})
	if err != nil {
		return err
	}

	_, err = svc.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(hashPointerKey),
		Body:   bytes.NewReader([]byte(md5Hash)),
	})
	if err != nil {
		return err
	}

	return nil
}
