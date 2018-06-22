// Copyright Jetstack Ltd. See LICENSE for details.
package amazon

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/hashicorp/terraform/terraform"
)

func (a *Amazon) RemoteStateName() string {
	return fmt.Sprintf(
		"%s%s-terraform-state",
		a.conf.Amazon.BucketPrefix,
		a.Region(),
	)
}

const DynamoDBKey = "LockID"

// TODO: remove me, deprecated
func (a *Amazon) RemoteStateBucketName() string {
	return a.RemoteStateName()
}

func (a *Amazon) RemoteState(namespace string, clusterName string, stackName string) string {
	return fmt.Sprintf(`terraform {
  backend "s3" {
    bucket = "%s"
    key = "%s"
    region = "%s"
    dynamodb_table ="%s"
  }
}`,
		a.RemoteStateName(),
		fmt.Sprintf("%s/%s/%s.tfstate", namespace, clusterName, stackName),
		a.Region(),
		a.RemoteStateName(),
	)
}

func (a *Amazon) RemoteStateBucketAvailable() (bool, error) {
	svc, err := a.S3()
	if err != nil {
		return false, err
	}

	_, err = svc.HeadBucket(&s3.HeadBucketInput{
		Bucket: aws.String(a.RemoteStateName()),
	})
	if err == nil {
		return true, nil
	} else if strings.HasPrefix(err.Error(), "NotFound:") {
		return false, nil
	}

	return false, fmt.Errorf("error while checking if remote state is available: %s", err)
}

func (a *Amazon) RemoteStateAvailable(bucketName string) (bool, error) {
	sess, err := a.Session()
	if err != nil {
		return false, fmt.Errorf("error getting session: %s", err)
	}

	svc := s3.New(sess)
	_, err = svc.HeadBucket(&s3.HeadBucketInput{
		Bucket: &bucketName,
	})
	if err == nil {
		return true, nil
	} else if strings.HasPrefix(err.Error(), "NotFound:") {
		return false, nil
	} else {
		return false, fmt.Errorf("error while checking if remote state is available: %s", err)
	}
}
func (a *Amazon) initRemoteStateBucket() error {
	svc, err := a.S3()
	if err != nil {
		return err
	}

	createBucketInput := &s3.CreateBucketInput{
		Bucket: aws.String(a.RemoteStateName()),
	}

	if a.Region() != "us-east-1" {
		createBucketInput.CreateBucketConfiguration = &s3.CreateBucketConfiguration{
			LocationConstraint: aws.String(a.Region()),
		}
	}

	_, err = svc.CreateBucket(createBucketInput)
	if err != nil {
		return err
	}

	_, err = svc.PutBucketVersioning(&s3.PutBucketVersioningInput{
		Bucket: aws.String(a.RemoteStateName()),
		VersioningConfiguration: &s3.VersioningConfiguration{
			Status: aws.String("Enabled"),
		},
	})
	return err
}

func (a *Amazon) validateRemoteStateBucket() error {
	svc, err := a.S3()
	if err != nil {
		return err
	}

	_, err = svc.HeadBucket(&s3.HeadBucketInput{
		Bucket: aws.String(a.RemoteStateName()),
	})
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			if awsErr.Code() == "NotFound" {
				return a.initRemoteStateBucket()
			}
		}
		return fmt.Errorf("error looking for terraform state bucket: %s", err)
	}

	location, err := svc.GetBucketLocation(&s3.GetBucketLocationInput{
		Bucket: aws.String(a.RemoteStateName()),
	})
	if err != nil {
		return err
	}

	var bucketRegion string
	if location.LocationConstraint == nil {
		bucketRegion = "us-east-1"
	} else {
		bucketRegion = *location.LocationConstraint
	}

	if myRegion := a.Region(); bucketRegion != myRegion {
		return fmt.Errorf("bucket region is wrong, actual: %s expected: %s", bucketRegion, myRegion)
	}

	versioning, err := svc.GetBucketVersioning(&s3.GetBucketVersioningInput{
		Bucket: aws.String(a.RemoteStateName()),
	})
	if err != nil {
		return err
	}
	if *versioning.Status != "Enabled" {
		a.log.Warnf("state bucket %s has versioning disabled", a.RemoteStateName())
	}

	return nil

}

func (a *Amazon) initRemoteStateDynamoDB() error {
	svc, err := a.DynamoDB()
	if err != nil {
		return err
	}

	_, err = svc.CreateTable(&dynamodb.CreateTableInput{
		TableName: aws.String(a.RemoteStateName()),
		AttributeDefinitions: []*dynamodb.AttributeDefinition{
			&dynamodb.AttributeDefinition{
				AttributeName: aws.String(DynamoDBKey),
				AttributeType: aws.String("S"),
			},
		},
		KeySchema: []*dynamodb.KeySchemaElement{
			&dynamodb.KeySchemaElement{
				AttributeName: aws.String(DynamoDBKey),
				KeyType:       aws.String(dynamodb.KeyTypeHash),
			},
		},
		ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
			ReadCapacityUnits:  aws.Int64(1),
			WriteCapacityUnits: aws.Int64(1),
		},
	})
	return err
}

func (a *Amazon) validateRemoteStateDynamoDB() error {
	svc, err := a.DynamoDB()
	if err != nil {
		return err
	}

	describeOut, err := svc.DescribeTable(&dynamodb.DescribeTableInput{
		TableName: aws.String(a.RemoteStateName()),
	})
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			if awsErr.Code() == "ResourceNotFoundException" {
				return a.initRemoteStateDynamoDB()
			}
		}
		return fmt.Errorf("error looking for terraform state dynamodb: %s", err)
	}

	attributeFound := false
	for _, params := range describeOut.Table.AttributeDefinitions {
		if *params.AttributeName == DynamoDBKey {
			attributeFound = true
		}
	}
	if !attributeFound {
		return fmt.Errorf("the DynamoDB table '%s' doesn't contain a parameter named '%s'", a.RemoteStateName(), DynamoDBKey)
	}

	return nil
}

func (a *Amazon) RemoteStateData(path string) (*terraform.State, error) {
	downloader := s3manager.NewDownloader(a.session)
	buff := new(aws.WriteAtBuffer)
	bucketName := a.RemoteStateName()

	_, err := downloader.Download(buff,
		&s3.GetObjectInput{
			Bucket: &bucketName,
			Key:    &path,
		})
	if err != nil {
		return nil, fmt.Errorf("failed to read terraform state file '%s': %v", path, err)
	}

	state, err := terraform.ReadStateV3(buff.Bytes())
	if err != nil {
		return nil, fmt.Errorf("failed to decode remote terraform state '%s': %v", path, err)
	}

	return state, nil
}
