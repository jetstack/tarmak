package aws

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/go-multierror"
	vault "github.com/hashicorp/vault/api"
	"github.com/mitchellh/go-homedir"

	"github.com/jetstack/tarmak/pkg/tarmak/config"
	"github.com/jetstack/tarmak/pkg/tarmak/interfaces"
)

type AWS struct {
	conf *config.AWSConfig

	environment interfaces.Environment

	session *session.Session
	ec2     EC2
	s3      S3
	log     *logrus.Entry
}

type S3 interface {
	HeadBucket(input *s3.HeadBucketInput) (*s3.HeadBucketOutput, error)
}

type EC2 interface {
	DescribeInstances(input *ec2.DescribeInstancesInput) (*ec2.DescribeInstancesOutput, error)
	DescribeKeyPairs(input *ec2.DescribeKeyPairsInput) (*ec2.DescribeKeyPairsOutput, error)
	DescribeAvailabilityZones(input *ec2.DescribeAvailabilityZonesInput) (*ec2.DescribeAvailabilityZonesOutput, error)
}

var _ interfaces.Provider = &AWS{}

func NewFromConfig(environment interfaces.Environment, conf *config.AWSConfig) (*AWS, error) {
	a := &AWS{
		conf:        conf,
		environment: environment,
		log:         environment.Tarmak().Log(),
	}

	return a, nil
}

func (a *AWS) Name() string {
	return config.ProviderNameAWS
}

func (a *AWS) Region() string {
	return a.conf.Region
}

func (a *AWS) EC2() (EC2, error) {
	if a.ec2 == nil {
		sess, err := a.Session()
		if err != nil {
			return nil, fmt.Errorf("error getting AWS session: %s", err)
		}
		a.ec2 = ec2.New(sess)
	}
	return a.ec2, nil
}

func (a *AWS) S3() (S3, error) {
	if a.s3 == nil {
		sess, err := a.Session()
		if err != nil {
			return nil, fmt.Errorf("error getting AWS session: %s", err)
		}
		a.s3 = s3.New(sess)
	}
	return a.s3, nil
}

func (a *AWS) RemoteStateBucketName() string {
	return fmt.Sprintf(
		"%s%s-%s-terraform-state",
		a.environment.BucketPrefix(),
		a.environment.Name(),
		a.Region(),
	)
}

func (a *AWS) RemoteState(contextName, stackName string) string {
	return fmt.Sprintf(`terraform {
  backend "s3" {
    bucket = "%s"
    key = "%s"
    region = "%s"
    lock_table ="%s"
  }
}`,
		a.RemoteStateBucketName(),
		fmt.Sprintf("%s/%s/%s.tfstate", a.environment.Name(), contextName, stackName),
		a.Region(),
		a.RemoteStateBucketName(),
	)
}

func (a *AWS) RemoteStateBucketAvailable() (bool, error) {
	svc, err := a.S3()
	if err != nil {
		return false, err
	}

	_, err = svc.HeadBucket(&s3.HeadBucketInput{
		Bucket: aws.String(a.RemoteStateBucketName()),
	})
	if err == nil {
		return true, nil
	} else if strings.HasPrefix(err.Error(), "NotFound:") {
		return false, nil
	} else {
		return false, fmt.Errorf("error while checking if remote state is available: %s", err)
	}

	return false, nil
}

func (a *AWS) Variables() map[string]interface{} {
	output := map[string]interface{}{}
	if a.conf.KeyName != "" {
		output["key_name"] = a.conf.KeyName
	}
	if len(a.conf.AllowedAccountIDs) > 0 {
		output["allowed_account_ids"] = a.conf.AllowedAccountIDs
	}
	output["availability_zones"] = a.conf.AvailabiltyZones
	output["region"] = a.conf.Region

	return output
}

// This will return necessary environment variables
func (a *AWS) Environment() ([]string, error) {
	sess, err := a.Session()
	if err != nil {
		return []string{}, fmt.Errorf("error getting session: %s", err)
	}

	creds, err := sess.Config.Credentials.Get()
	if err != nil {
		return []string{}, fmt.Errorf("error getting credentials: %s", err)
	}

	return []string{
		fmt.Sprintf("AWS_ACCESS_KEY_ID=%s", creds.AccessKeyID),
		fmt.Sprintf("AWS_SECRET_ACCESS_KEY=%s", creds.SecretAccessKey),
		fmt.Sprintf("AWS_SESSION_TOKEN=%s", creds.SessionToken),
		fmt.Sprintf("AWS_DEFAULT_REGION=%s", a.Region),
	}, nil
}

// This reads the vault token from ~/.vault-token
func readVaultToken() (string, error) {
	homeDir, err := homedir.Dir()
	if err != nil {
		return "", err
	}

	filePath := filepath.Join(homeDir, ".vault-token")

	vaultToken, err := ioutil.ReadFile(filePath)
	if err != nil {
		return "", err
	}

	return string(vaultToken), nil
}

func (a *AWS) Validate() error {
	sess, err := a.Session()
	if err != nil {
		return fmt.Errorf("error getting AWS session: %s", err)
	}

	svc := ec2.New(sess)
	err = a.validateAvailabilityZones(svc)
	if err != nil {
		return err
	}

	return nil

}

func (a *AWS) validateAvailabilityZones(svc *ec2.EC2) error {
	var result error

	zones, err := svc.DescribeAvailabilityZones(&ec2.DescribeAvailabilityZonesInput{
		Filters: []*ec2.Filter{
			&ec2.Filter{
				Name:   aws.String("state"),
				Values: []*string{aws.String("available")},
			},
		},
	})

	if err != nil {
		return err
	}

	if len(zones.AvailabilityZones) == 0 {
		return fmt.Errorf(
			"no availability zone found for region '%s'",
			a.Region,
		)
	}

	for _, zoneConfigured := range a.conf.AvailabiltyZones {
		found := false
		for _, zone := range zones.AvailabilityZones {
			if zone.ZoneName != nil && *zone.ZoneName == zoneConfigured {
				found = true
				break
			}
		}
		if !found {
			result = multierror.Append(result, fmt.Errorf(
				"specified invalid availability zone '%s' for region '%s'",
				zoneConfigured,
				a.Region,
			))
		}
	}
	if result != nil {
		return result
	}

	if len(a.conf.AvailabiltyZones) == 0 {
		zone := zones.AvailabilityZones[0].ZoneName
		if zone == nil {
			return fmt.Errorf("error determining availabilty zone")
		}
		a.log.Debugf("no availability zones specified selecting zone: %s", *zone)
		a.conf.AvailabiltyZones = []string{*zone}
	}

	return nil
}

func (a *AWS) Session() (*session.Session, error) {

	// return cached session
	if a.session != nil {
		return a.session, nil
	}

	// use default config, if vault disabled
	if a.conf.VaultPath != "" {
		sess, err := a.vaultSession()
		if err != nil {
			return nil, err
		} else {
			a.session = sess
			return a.session, nil
		}
	}

	a.session = session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	a.session.Config.Region = aws.String(a.Region())
	return a.session, nil
}

func (a *AWS) vaultSession() (*session.Session, error) {
	vaultClient, err := vault.NewClient(nil)
	if err != nil {
		return nil, err
	}

	// without vault token lookup vault token file
	if os.Getenv("VAULT_TOKEN") == "" {
		vaultToken, err := readVaultToken()
		if err != nil {
			a.log.Debug("failed to read vault token file: ", err)
		} else {
			vaultClient.SetToken(vaultToken)
		}
	}

	awsSecret, err := vaultClient.Logical().Read(a.conf.VaultPath)
	if err != nil {
		return nil, err
	}
	if awsSecret == nil || awsSecret.Data == nil {
		return nil, fmt.Errorf("vault did not return data at path '%s'", a.conf.VaultPath)
	}

	values := []string{}

	for _, key := range []string{"access_key", "secret_key", "security_token"} {
		val, ok := awsSecret.Data[key]
		if !ok {
			return nil, fmt.Errorf("vault did not return data with key '%s'", key)
		}
		valString, ok := val.(string)
		if !ok {
			return nil, fmt.Errorf("vault did not return data with a string in key '%s'", key)
		}
		values = append(values, valString)
	}

	creds := credentials.NewStaticCredentials(values[0], values[1], values[2])

	sess := session.Must(session.NewSession())
	sess.Config.Region = aws.String(a.Region())
	sess.Config.Credentials = creds

	return sess, nil
}

func (a *AWS) validateAWSKeyPair(svc *ec2.EC2) error {
	if a.conf.KeyName == "" {
		a.conf.KeyName = fmt.Sprintf("tarmak_%s", a.environment.Name())
	}

	keypairs, err := svc.DescribeKeyPairs(&ec2.DescribeKeyPairsInput{
		KeyNames: []*string{aws.String(a.conf.KeyName)},
	})
	if err != nil {
		return err
	}

	var awsKeyPair *ec2.KeyPairInfo
	if len(keypairs.KeyPairs) == 0 {
		// TODO create me
	}
	if len(keypairs.KeyPairs) != 1 {
		return fmt.Errorf("unexpected number of keypairs found: %d", len(keypairs.KeyPairs))
	} else {
		awsKeyPair = keypairs.KeyPairs[0]
	}

	// TODO implement me properly
	if *awsKeyPair.KeyFingerprint != "" {
		return fmt.Errorf("aws key pair is not matching the local one")
	}

	// read private key from disk

	// if local key path + pub key is given => use that
	// maybe check if the two keys match (opt)

	// if local private key path given, make sure its unencrypted and upload to tarmak_<context_name>

	// if nothing of the two is there generate key pair in the well know path (if not exist and upload that if not exsist then verify as before:w

	return nil

}

func (a *AWS) RemoteStateAvailable(bucketName string) (bool, error) {
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
