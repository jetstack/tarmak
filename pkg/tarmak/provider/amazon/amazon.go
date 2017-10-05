package amazon

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"

	"github.com/Sirupsen/logrus"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/go-multierror"
	vault "github.com/hashicorp/vault/api"

	clusterv1alpha1 "github.com/jetstack/tarmak/pkg/apis/cluster/v1alpha1"
	tarmakv1alpha1 "github.com/jetstack/tarmak/pkg/apis/tarmak/v1alpha1"
	"github.com/jetstack/tarmak/pkg/tarmak/interfaces"
	"github.com/jetstack/tarmak/pkg/tarmak/utils/input"
)

var _ interfaces.Provider = &Amazon{}

type Amazon struct {
	conf *tarmakv1alpha1.Provider

	tarmak interfaces.Tarmak

	availabilityZones *[]string

	session  *session.Session
	ec2      EC2
	s3       S3
	dynamodb DynamoDB
	route53  Route53
	log      *logrus.Entry
}

type S3 interface {
	HeadBucket(input *s3.HeadBucketInput) (*s3.HeadBucketOutput, error)
	CreateBucket(input *s3.CreateBucketInput) (*s3.CreateBucketOutput, error)
	GetBucketVersioning(input *s3.GetBucketVersioningInput) (*s3.GetBucketVersioningOutput, error)
	GetBucketLocation(input *s3.GetBucketLocationInput) (*s3.GetBucketLocationOutput, error)
	PutBucketVersioning(input *s3.PutBucketVersioningInput) (*s3.PutBucketVersioningOutput, error)
}

type EC2 interface {
	DescribeInstances(input *ec2.DescribeInstancesInput) (*ec2.DescribeInstancesOutput, error)
	ImportKeyPair(input *ec2.ImportKeyPairInput) (*ec2.ImportKeyPairOutput, error)
	DescribeKeyPairs(input *ec2.DescribeKeyPairsInput) (*ec2.DescribeKeyPairsOutput, error)
	DescribeAvailabilityZones(input *ec2.DescribeAvailabilityZonesInput) (*ec2.DescribeAvailabilityZonesOutput, error)
	DescribeRegions(input *ec2.DescribeRegionsInput) (*ec2.DescribeRegionsOutput, error)
}

type DynamoDB interface {
	DescribeTable(input *dynamodb.DescribeTableInput) (*dynamodb.DescribeTableOutput, error)
	CreateTable(input *dynamodb.CreateTableInput) (*dynamodb.CreateTableOutput, error)
}

type Route53 interface {
	CreateHostedZone(input *route53.CreateHostedZoneInput) (*route53.CreateHostedZoneOutput, error)
	GetHostedZone(input *route53.GetHostedZoneInput) (*route53.GetHostedZoneOutput, error)
	ListHostedZonesByName(input *route53.ListHostedZonesByNameInput) (*route53.ListHostedZonesByNameOutput, error)
}

var _ interfaces.Provider = &Amazon{}

func NewFromConfig(tarmak interfaces.Tarmak, conf *tarmakv1alpha1.Provider) (*Amazon, error) {

	a := &Amazon{
		conf:   conf,
		log:    tarmak.Log().WithField("provider_name", conf.ObjectMeta.Name),
		tarmak: tarmak,
	}

	return a, nil
}

func (a *Amazon) Name() string {
	return a.conf.Name
}

func (a *Amazon) Cloud() string {
	return clusterv1alpha1.CloudAmazon
}

// This parameters should include non sensitive information to identify a provider
func (a *Amazon) Parameters() map[string]string {
	p := map[string]string{
		"name":          a.Name(),
		"cloud":         a.Cloud(),
		"public_zone":   a.conf.Amazon.PublicZone,
		"bucket_prefix": a.conf.Amazon.BucketPrefix,
	}
	if a.conf.Amazon.VaultPath != "" {
		p["vault_path"] = a.conf.Amazon.VaultPath
	}
	if a.conf.Amazon.Profile != "" {
		p["amazon_profile"] = a.conf.Amazon.Profile
	}
	return p
}

func (a *Amazon) String() string {
	return fmt.Sprintf("%s[%s]", a.Cloud(), a.Name())
}

func (a *Amazon) ListRegions() (regions []string, err error) {
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

func (a *Amazon) AskEnvironmentLocation(init interfaces.Initialize) (location string, err error) {
	regions, err := a.ListRegions()
	if err != nil {
		return "", err
	}

	regionPos, err := init.Input().AskSelection(&input.AskSelection{
		Query:   "In which region should this environment reside?",
		Choices: regions,
		Default: -1,
	})
	if err != nil {
		return "", err
	}

	return regions[regionPos], nil
}

func (a *Amazon) AskInstancePoolLocation(init interfaces.Initialize) (zones []string, err error) {

	zones, err = a.getAvailablityZoneByRegion()
	if err != nil {
		return []string{}, fmt.Errorf("failed to get availabilty zones: %v", err)
	}

	if len(zones) == 0 {
		return []string{}, fmt.Errorf("no availability zones found for region '%s'", a.Region())
	}

	sChoices := make([]bool, len(zones))
	sChoices[0] = true

	multiSel := &input.AskMultipleSelection{
		AskSelection: &input.AskSelection{
			Query:   "Please select availabilty zones. Enter numbers to toggle selection.",
			Choices: zones,
			Default: 1,
		},
		SelectedChoices: sChoices,
		MinSelected:     1,
		MaxSelected:     len(zones),
	}

	return init.Input().AskMultipleSelection(multiSel)
}

func (a *Amazon) Region() string {
	// without environment selected, fall back to default region
	if a.tarmak.Environment() == nil {
		return "us-east-1"
	}
	return a.tarmak.Environment().Location()
}

// This return the availabililty zones that are used for a cluster
func (a *Amazon) AvailabilityZones() (availabiltyZones []string) {
	if a.availabilityZones != nil {
		return *a.availabilityZones
	}

	subnets := a.tarmak.Cluster().Subnets()
	zones := make(map[string]bool)

	for _, subnet := range subnets {
		zones[subnet.Zone] = true
	}

	a.availabilityZones = &availabiltyZones

	for zone, _ := range zones {
		availabiltyZones = append(availabiltyZones, zone)
	}

	sort.Strings(availabiltyZones)

	return availabiltyZones
}

func (a *Amazon) EC2() (EC2, error) {
	if a.ec2 == nil {
		sess, err := a.Session()
		if err != nil {
			return nil, fmt.Errorf("error getting Amazon session: %s", err)
		}
		a.ec2 = ec2.New(sess)
	}
	return a.ec2, nil
}

func (a *Amazon) S3() (S3, error) {
	if a.s3 == nil {
		sess, err := a.Session()
		if err != nil {
			return nil, fmt.Errorf("error getting Amazon session: %s", err)
		}
		a.s3 = s3.New(sess)
	}
	return a.s3, nil
}

func (a *Amazon) DynamoDB() (DynamoDB, error) {
	if a.dynamodb == nil {
		sess, err := a.Session()
		if err != nil {
			return nil, fmt.Errorf("error getting Amazon session: %s", err)
		}
		a.dynamodb = dynamodb.New(sess)
	}
	return a.dynamodb, nil
}

func (a *Amazon) Route53() (Route53, error) {
	if a.route53 == nil {
		sess, err := a.Session()
		if err != nil {
			return nil, fmt.Errorf("error getting Amazon session: %s", err)
		}
		a.route53 = route53.New(sess)
	}
	return a.route53, nil
}

func (a *Amazon) Variables() map[string]interface{} {
	output := map[string]interface{}{}
	output["key_name"] = a.KeyName()
	if len(a.conf.Amazon.AllowedAccountIDs) > 0 {
		output["allowed_account_ids"] = a.conf.Amazon.AllowedAccountIDs
	}
	output["availability_zones"] = a.AvailabilityZones()
	output["region"] = a.Region()

	output["public_zone"] = a.conf.Amazon.PublicZone
	output["public_zone_id"] = a.conf.Amazon.PublicHostedZoneID
	output["bucket_prefix"] = a.conf.Amazon.BucketPrefix

	return output
}

// This will return necessary environment variables
func (a *Amazon) Environment() ([]string, error) {
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
func (a *Amazon) readVaultToken() (string, error) {
	homeDir := a.tarmak.HomeDir()

	filePath := filepath.Join(homeDir, ".vault-token")

	vaultToken, err := ioutil.ReadFile(filePath)
	if err != nil {
		return "", err
	}

	return string(vaultToken), nil
}

func (a *Amazon) Validate() error {
	var result error
	var err error

	// These checks only make sense with an environment given
	if a.tarmak.Environment() != nil {
		err = a.validateRemoteStateBucket()
		if err != nil {
			result = multierror.Append(result, err)
		}

		err = a.validateRemoteStateDynamoDB()
		if err != nil {
			result = multierror.Append(result, err)
		}

		err = a.validateAvailabilityZones()
		if err != nil {
			result = multierror.Append(result, err)
		}

		err = a.validateAWSKeyPair()
		if err != nil {
			result = multierror.Append(result, err)
		}

	}

	err = a.validatePublicZone()
	if err != nil {
		result = multierror.Append(result, err)
	}

	if result != nil {
		return result
	}
	return nil

}

func (a *Amazon) getAvailablityZoneByRegion() (zones []string, err error) {
	svc, err := a.EC2()
	if err != nil {
		return []string{}, fmt.Errorf("error getting AWS EC2 session: %s", err)
	}

	ec2Zones, err := svc.DescribeAvailabilityZones(&ec2.DescribeAvailabilityZonesInput{
		Filters: []*ec2.Filter{
			&ec2.Filter{
				Name:   aws.String("state"),
				Values: []*string{aws.String("available")},
			},
		},
	})
	if err != nil {
		return []string{}, err
	}

	for _, zone := range ec2Zones.AvailabilityZones {
		zones = append(zones, *zone.ZoneName)
	}

	sort.Strings(zones)

	return zones, nil
}

func (a *Amazon) validateAvailabilityZones() error {
	var result error

	zones, err := a.getAvailablityZoneByRegion()
	if err != nil {
		return err
	}

	if len(zones) == 0 {
		return fmt.Errorf(
			"no availability zone found for region '%s'",
			a.Region(),
		)
	}

	availabilityZones := a.AvailabilityZones()

	for _, zoneConfigured := range availabilityZones {
		found := false
		for _, zone := range zones {
			if zone != "" && zone == zoneConfigured {
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
		zone := zones[0]
		if zone == "" {
			return fmt.Errorf("error determining availabilty zone")
		}
		a.log.Debugf("no availability zones specified selecting zone: %s", zone)
		availabilityZones = []string{zone}
		a.availabilityZones = &availabilityZones
	}

	return nil
}

func (a *Amazon) Session() (*session.Session, error) {

	// return cached session
	if a.session != nil {
		return a.session, nil
	}

	// use default config, if vault disabled
	if a.conf.Amazon.VaultPath != "" {
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
		Profile:           a.conf.Amazon.Profile,
	}))
	a.session.Config.Region = aws.String(a.Region())
	return a.session, nil
}

func (a *Amazon) vaultSession() (*session.Session, error) {
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

	awsSecret, err := vaultClient.Logical().Read(a.conf.Amazon.VaultPath)
	if err != nil {
		return nil, err
	}
	if awsSecret == nil || awsSecret.Data == nil {
		return nil, fmt.Errorf("vault did not return data at path '%s'", a.conf.Amazon.VaultPath)
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
func (a *Amazon) InstanceType(typeIn string) (typeOut string, err error) {
	if typeIn == clusterv1alpha1.InstancePoolSizeTiny {
		return "t2.nano", nil
	}
	if typeIn == clusterv1alpha1.InstancePoolSizeSmall {
		return "t2.medium", nil
	}
	if typeIn == clusterv1alpha1.InstancePoolSizeMedium {
		return "m4.large", nil
	}
	if typeIn == clusterv1alpha1.InstancePoolSizeLarge {
		return "m4.xlarge", nil
	}

	// TODO: Validate custom instance type here
	return typeIn, nil
}

// This methods converts and possibly validates a generic volume type to a
// provider specifc
func (a *Amazon) VolumeType(typeIn string) (typeOut string, err error) {
	if typeIn == clusterv1alpha1.VolumeTypeHDD {
		return "st2", nil
	}
	if typeIn == clusterv1alpha1.VolumeTypeSSD {
		return "gp2", nil
	}
	// TODO: Validate custom instance type here
	return typeIn, nil
}
