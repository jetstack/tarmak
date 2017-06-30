package config

type GCPConfig struct {
	ProjectName string `yaml:"projectName,omitempty"`
	AccountName string `yaml:"accountName,omitempty"`
}
