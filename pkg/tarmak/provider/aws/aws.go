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

	clusterv1alpha1 "github.com/jetstack/tarmak/pkg/apis/cluster/v1alpha1"
	tarmakv1alpha1 "github.com/jetstack/tarmak/pkg/apis/tarmak/v1alpha1"
	"github.com/jetstack/tarmak/pkg/tarmak/interfaces"
)

type AWS struct {
	conf *tarmakv1alpha1.Provider

	tarmak interfaces.Tarmak

	availabilityZones *[]string

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
	ImportKeyPair(input *ec2.ImportKeyPairInput) (*ec2.ImportKeyPairOutput, error)
	DescribeKeyPairs(input *ec2.DescribeKeyPairsInput) (*ec2.DescribeKeyPairsOutput, error)
	DescribeAvailabilityZones(input *ec2.DescribeAvailabilityZonesInput) (*ec2.DescribeAvailabilityZonesOutput, error)
	DescribeRegions(input *ec2.DescribeRegionsInput) (*ec2.DescribeRegionsOutput, error)
}

var _ interfaces.Provider = &AWS{}

func NewFromConfig(tarmak interfaces.Tarmak, conf *tarmakv1alpha1.Provider) (*AWS, error) {

	a := &AWS{
		conf:   conf,
		log:    tarmak.Log().WithField("provider_name", conf.ObjectMeta.Name),
		tarmak: tarmak,
	}

	return a, nil
}

func (a *AWS) Name() string {
	return a.conf.Name
}

func (a *AWS) Cloud() string {
	return clusterv1alpha1.CloudAmazon
}

// This parameters should include non sensitive information to identify a provider
func (a *AWS) Parameters() map[string]string {
	p := map[string]string{
		"public_zone":   a.conf.AWS.PublicZone,
		"bucket_prefix": a.conf.AWS.BucketPrefix,
	}
	if a.conf.AWS.VaultPath != "" {
		p["vault_path"] = a.conf.AWS.VaultPath
	}
	if a.conf.AWS.Profile != "" {
		p["aws_profile"] = a.conf.AWS.Profile
	}
	return p
}

func (a *AWS) ListRegions() (regions []string, err error) {
	svc, err := a.EC2()
	if err != nil {
		return regions, err
	}

	regionsOutput, err := svc.DescribeRegions(&ec2.DescribeRegionsInput{})
	if err != nil {
		return regions, err
	}

	for _, region := range regionsOutput.Regions {
		regions = append(regions, *region.RegionName)
	}

	return regions, nil

}

func (a *AWS) Region() string {
	return a.tarmak.Context().Region()
}

// This return the availabililty zones that are used for a cluster
func (a *AWS) AvailabilityZones() (availabiltyZones []string) {
	if a.availabilityZones != nil {
		return *a.availabilityZones
	}

	subnets := a.tarmak.Context().Subnets()
	zones := make(map[string]bool)

	for _, subnet := range subnets {
		zones[subnet.Zone] = true
	}

	a.availabilityZones = &availabiltyZones

	for zone, _ := range zones {
		availabiltyZones = append(availabiltyZones, zone)
	}

	return availabiltyZones
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
		"%s-%s-terraform-state",
		a.conf.AWS.BucketPrefix,
		a.Region(),
	)
}

func (a *AWS) RemoteState(namespace string, clusterName string, stackName string) string {
	return fmt.Sprintf(`terraform {
  backend "s3" {
    bucket = "%s"
    key = "%s"
    region = "%s"
    lock_table ="%s"
  }
}`,
		a.RemoteStateBucketName(),
		fmt.Sprintf("%s/%s/%s.tfstate", namespace, clusterName, stackName),
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
	}

	return false, fmt.Errorf("error while checking if remote state is available: %s", err)
}

func (a *AWS) Variables() map[string]interface{} {
	output := map[string]interface{}{}
	output["key_name"] = a.KeyName()
	if len(a.conf.AWS.AllowedAccountIDs) > 0 {
		output["allowed_account_ids"] = a.conf.AWS.AllowedAccountIDs
	}
	output["availability_zones"] = a.AvailabilityZones()
	output["region"] = a.Region()

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
		fmt.Sprintf("AWS_DEFAULT_REGION=%s", a.Region()),
	}, nil
}

// This reads the vault token from ~/.vault-token
func (a *AWS) readVaultToken() (string, error) {
	homeDir := a.tarmak.HomeDir()

	filePath := filepath.Join(homeDir, ".vault-token")

	vaultToken, err := ioutil.ReadFile(filePath)
	if err != nil {
		return "", err
	}

	return string(vaultToken), nil
}

func (a *AWS) Validate() error {
	err := a.validateAvailabilityZones()
	if err != nil {
		return err
	}

	err = a.validateAWSKeyPair()
	if err != nil {
		return err
	}

	return nil

}

func (a *AWS) validateAvailabilityZones() error {
	var result error

	svc, err := a.EC2()
	if err != nil {
		return fmt.Errorf("error getting AWS EC2 session: %s", err)
	}

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
			a.Region(),
		)
	}

	availabilityZones := a.AvailabilityZones()

	for _, zoneConfigured := range availabilityZones {
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
				a.Region(),
			))
		}
	}
	if result != nil {
		return result
	}

	if len(availabilityZones) == 0 {
		zone := zones.AvailabilityZones[0].ZoneName
		if zone == nil {
			return fmt.Errorf("error determining availabilty zone")
		}
		a.log.Debugf("no availability zones specified selecting zone: %s", *zone)
		availabilityZones = []string{*zone}
		a.availabilityZones = &availabilityZones
	}

	return nil
}

func (a *AWS) Session() (*session.Session, error) {

	// return cached session
	if a.session != nil {
		return a.session, nil
	}

	// use default config, if vault disabled
	if a.conf.AWS.VaultPath != "" {
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
		vaultToken, err := a.readVaultToken()
		if err != nil {
			a.log.Debug("failed to read vault token file: ", err)
		} else {
			vaultClient.SetToken(vaultToken)
		}
	}

	awsSecret, err := vaultClient.Logical().Read(a.conf.AWS.VaultPath)
	if err != nil {
		return nil, err
	}
	if awsSecret == nil || awsSecret.Data == nil {
		return nil, fmt.Errorf("vault did not return data at path '%s'", a.conf.AWS.VaultPath)
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

// This methods converts and possibly validates a generic instance type to a
// provider specifc
func (a *AWS) InstanceType(typeIn string) (typeOut string, err error) {
	if typeIn == clusterv1alpha1.ServerPoolSizeTiny {
		return "t2.nano", nil
	}
	if typeIn == clusterv1alpha1.ServerPoolSizeSmall {
		return "m3.medium", nil
	}
	if typeIn == clusterv1alpha1.ServerPoolSizeMedium {
		return "m4.large", nil
	}
	if typeIn == clusterv1alpha1.ServerPoolSizeLarge {
		return "m4.xlarge", nil
	}

	// TODO: Validate custom instance type here
	return typeIn, nil
}

// This methods converts and possibly validates a generic volume type to a
// provider specifc
func (a *AWS) VolumeType(typeIn string) (typeOut string, err error) {
	if typeIn == clusterv1alpha1.VolumeTypeHDD {
		return "st2", nil
	}
	if typeIn == clusterv1alpha1.VolumeTypeSSD {
		return "gp2", nil
	}
	// TODO: Validate custom instance type here
	return typeIn, nil
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
