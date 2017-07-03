package config

import (
	"fmt"
	"io"
	"os"
	"strings"

	logrus "github.com/Sirupsen/logrus"
	"github.com/hashicorp/go-multierror"
)

const (
	StackNameState      = "state"
	StackNameNetwork    = "network"
	StackNameTools      = "tools"
	StackNameVault      = "vault"
	StackNameKubernetes = "kubernetes"
	ProviderNameAWS     = "aws"
	ProviderNameGCP     = "gcp"
)

type Config struct {
	CurrentContext string        `yaml:"currentContext,omitempty"` // <environmentName>-<contextName>
	Environments   []Environment `yaml:"environments,omitempty"`

	Contact string `yaml:"contact,omitempty"`
	Project string `yaml:"project,omitempty"`
}

type Tarmak interface {
	Log() *logrus.Entry
	RootPath() string
	Context() *Context
}

type File interface {
	io.Closer
	io.Reader
	io.Seeker
	Readdir(count int) ([]os.FileInfo, error)
	Stat() (os.FileInfo, error)
}

func (c *Config) GetContext() (*Context, error) {
	for _, environment := range c.Environments {
		if !strings.HasPrefix(c.CurrentContext, fmt.Sprintf("%s-", environment.Name)) {
			continue
		}
		for _, context := range environment.Contexts {
			if context.GetName() == c.CurrentContext {
				return &context, nil
			}
		}
	}
	return nil, fmt.Errorf("context '%s' not found", c.CurrentContext)
}

func (c *Config) Validate() error {
	var result error

	for posEnvironment, _ := range c.Environments {
		// set my link onto environment
		c.Environments[posEnvironment].config = c

		// validate environment
		if err := c.Environments[posEnvironment].Validate(); err != nil {
			result = multierror.Append(result, err)
		}
	}

	return result
}

func (c *Config) TerraformVars() map[string]interface{} {
	return map[string]interface{}{
		"project": c.Project,
		"contact": c.Contact,
	}
}

func DefaultConfigSingle() *Config {
	return &Config{
		CurrentContext: "devsingle-cluster",
		Project:        "tarmak-dev",
		Contact:        "tech@jetstack.io",
		Environments: []Environment{
			Environment{
				AWS: &AWSConfig{
					VaultPath: "jetstack/aws/jetstack-dev/sts/admin",
					Region:    "eu-west-1",
					KeyName:   "jetstack_nonprod",
				},
				Name: "devsingle",
				Contexts: []Context{
					Context{
						Name: "cluster",
						Stacks: []Stack{
							Stack{
								State: &StackState{
									BucketPrefix: "jetstack-tarmak-",
									PublicZone:   "devsingle.dev.tarmak.io",
								},
							},
							Stack{
								Network: &StackNetwork{
									NetworkCIDR: "10.98.0.0/20",
									PrivateZone: "devsingle.dev.tarmak.local",
								},
							},
							Stack{
								Tools: &StackTools{},
							},
							Stack{
								Vault: &StackVault{},
							},
							Stack{
								Kubernetes: &StackKubernetes{
									EtcdCount:       3,
									WorkerSpotPrice: 0.035,
								},
							},
						},
					},
				},
			},
		},
	}
}

func DefaultConfigHub() *Config {
	return &Config{
		CurrentContext: "devmulti-hub",
		Project:        "tarmak-dev",
		Contact:        "tech@jetstack.io",
		Environments: []Environment{
			Environment{
				AWS: &AWSConfig{
					VaultPath: "jetstack/aws/jetstack-dev/sts/admin",
					Region:    "eu-west-1",
					KeyName:   "jetstack_nonprod",
				},
				Name: "devmulti",
				Contexts: []Context{
					Context{
						Name: "hub",
						Stacks: []Stack{
							Stack{
								State: &StackState{
									BucketPrefix: "jetstack-tarmak-",
									PublicZone:   "devmulti.dev.tarmak.io",
								},
							},
							Stack{
								Network: &StackNetwork{
									NetworkCIDR: "10.99.0.0/20",
									PrivateZone: "devmulti.dev.tarmak.local",
								},
							},
							Stack{
								Tools: &StackTools{},
							},
							Stack{
								Vault: &StackVault{},
							},
						},
					},
					Context{
						Name: "cluster",
						Stacks: []Stack{
							Stack{
								Network: &StackNetwork{
									NetworkCIDR: "10.99.16.0/20",
									PeerContext: "hub",
								},
							},
							Stack{
								Kubernetes: &StackKubernetes{
									EtcdCount:       3,
									WorkerSpotPrice: 0.035,
								},
							},
						},
					},
				},
			},
		},
	}
}
