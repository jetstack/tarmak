package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	log "github.com/Sirupsen/logrus"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/hashicorp/go-multierror"
	vault "github.com/hashicorp/vault/api"
	"github.com/mitchellh/go-homedir"
)

type AWSConfig struct {
	VaultPath   string `yaml:"vaultPath,omitempty"`
	AccountID   string `yaml:"accountID,omitempty"`
	Region      string `yaml:"region,omitempty"`
	environment *[]string
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

// This will return necessary environment variables
func (a *AWSConfig) Environment() ([]string, error) {

	// check if metadata service is available, if yes return and use IAM role
	sess := session.Must(session.NewSession())
	ec2meta := ec2metadata.New(sess)
	if ec2meta.Available() {
		log.Infof("ec2metadata service available, use IAM credentials for AWS auth")
		return []string{}, nil
	}

	output := []string{}

	awsSecretMappings := []struct {
		Env   string
		Vault string
	}{
		{
			"AWS_ACCESS_KEY_ID",
			"access_key",
		},
		{
			"AWS_SECRET_ACCESS_KEY",
			"secret_key",
		},
		{
			"AWS_SESSION_TOKEN",
			"security_token",
		},
	}

	// if vault is configured get credentials from there
	if a.VaultPath != "" {
		vaultClient, err := vault.NewClient(nil)
		if err != nil {
			return output, err
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
			return output, err
		}
		if awsSecret == nil || awsSecret.Data == nil {
			return output, fmt.Errorf("vault did not return data at path '%s'", a.VaultPath)
		}

		var result error

		for _, mapping := range awsSecretMappings {
			val, ok := awsSecret.Data[mapping.Vault]
			if !ok {
				result = multierror.Append(result, fmt.Errorf("Vault response did not contain required field '%s'", mapping.Vault))
				continue
			}
			output = append(output, fmt.Sprintf("%s=%s", mapping.Env, val))
		}

		return output, result
	}

	// by default forward environment variables
	for _, mapping := range awsSecretMappings {
		if val := os.Getenv(mapping.Env); val != "" {
			output = append(output, fmt.Sprintf("%s=%s", mapping.Env, val))
		}
	}

	return output, nil

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

func (a *AWSConfig) RemoteStateAvailable(bucketName string) bool {
	return false
}
