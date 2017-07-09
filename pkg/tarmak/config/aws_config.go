package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/go-multierror"
	vault "github.com/hashicorp/vault/api"
	"github.com/mitchellh/go-homedir"
)

type AWSConfig struct {
	VaultPath         string   `yaml:"vaultPath,omitempty"`
	AllowedAccountIDs []string `yaml:"allowedAccountIDs,omitempty"`
	AvailabiltyZones  []string `yaml:"availabilityZones,omitempty"`
	Region            string   `yaml:"region,omitempty"`
	KeyName           string   `yaml:"keyName,omitempty"` // ec2 key pair name

	session *session.Session
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

func (a *AWSConfig) Validate() error {
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

func (a *AWSConfig) validateAvailabilityZones(svc *ec2.EC2) error {
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

	for _, zoneConfigured := range a.AvailabiltyZones {
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

	if len(a.AvailabiltyZones) == 0 {
		zone := zones.AvailabilityZones[0].ZoneName
		if zone == nil {
			return fmt.Errorf("error determining availabilty zone")
		}
		log.Debugf("No availability zones specified selecting zone: %s", zone)
		a.AvailabiltyZones = []string{*zone}
	}

	return nil
}

func (a *AWSConfig) Session() (*session.Session, error) {

	// return cached session
	if a.session != nil {
		return a.session, nil
	}

	// use default config, if vault disabled
	if a.VaultPath != "" {
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
	a.session.Config.Region = &a.Region
	return a.session, nil
}

func (a *AWSConfig) vaultSession() (*session.Session, error) {
	vaultClient, err := vault.NewClient(nil)
	if err != nil {
		return nil, err
	}

	// without vault token lookup vault token file
	if os.Getenv("VAULT_TOKEN") == "" {
		vaultToken, err := readVaultToken()
		if err != nil {
			log.Debug("failed to read vault token file: ", err)
		} else {
			vaultClient.SetToken(vaultToken)
		}
	}

	awsSecret, err := vaultClient.Logical().Read(a.VaultPath)
	if err != nil {
		return nil, err
	}
	if awsSecret == nil || awsSecret.Data == nil {
		return nil, fmt.Errorf("vault did not return data at path '%s'", a.VaultPath)
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
	sess.Config.Region = &a.Region
	sess.Config.Credentials = creds

	return sess, nil
}

// This will return necessary environment variables
func (a *AWSConfig) Environment() ([]string, error) {
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

func (a *AWSConfig) RemoteState(bucketName, environmentName, contextName, stackName string) string {
	return fmt.Sprintf(`terraform {
  backend "s3" {
    bucket = "%s"
    key = "%s"
    region = "%s"
    lock_table ="%s"
  }
}`,
		bucketName,
		fmt.Sprintf("%s/%s/%s.tfstate", environmentName, contextName, stackName),
		a.Region,
		bucketName,
	)
}

func (a *AWSConfig) TerraformVars() map[string]interface{} {
	output := map[string]interface{}{}
	if a.KeyName != "" {
		output["key_name"] = a.KeyName
	}
	if len(a.AllowedAccountIDs) > 0 {
		output["allowed_account_ids"] = a.AllowedAccountIDs
	}
	output["availability_zones"] = a.AvailabiltyZones
	output["region"] = a.Region

	return output
}

func (a *AWSConfig) RemoteStateAvailable(bucketName string) (bool, error) {
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
