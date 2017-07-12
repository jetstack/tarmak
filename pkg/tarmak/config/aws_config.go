package config

import ()

type AWSConfig struct {
	VaultPath         string   `yaml:"vaultPath,omitempty"`
	AllowedAccountIDs []string `yaml:"allowedAccountIDs,omitempty"`
	AvailabiltyZones  []string `yaml:"availabilityZones,omitempty"`
	Region            string   `yaml:"region,omitempty"`
	KeyName           string   `yaml:"keyName,omitempty"` // ec2 key pair name
}
