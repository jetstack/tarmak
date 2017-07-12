package config

import ()

type Environment struct {
	Name string     `yaml:"name,omitempty"` // only alphanumeric lowercase
	AWS  *AWSConfig `yaml:"aws,omitempty"`
	GCP  *GCPConfig `yaml:"gcp,omitempty"`

	Contact string `yaml:"contact,omitempty"`
	Project string `yaml:"project,omitempty"`

	SSHKeyPath string `yaml:"sshKeyPath,omitempty"`

	Contexts []Context `yaml:"contexts,omitempty"`
}
