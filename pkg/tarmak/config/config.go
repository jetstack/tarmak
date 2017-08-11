package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"

	"github.com/jetstack/tarmak/pkg/tarmak/interfaces"
	"github.com/jetstack/tarmak/pkg/tarmak/utils"
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

func configPath(t interfaces.Tarmak) string {
	return filepath.Join(t.ConfigPath(), "tarmak.yaml")
}

func ReadConfig(t interfaces.Tarmak) (*Config, error) {
	path := configPath(t)

	configBytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config Config
	err = yaml.Unmarshal(configBytes, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

func MergeEnvironment(t interfaces.Tarmak, in Environment) error {
	contextName := fmt.Sprintf("%s-%s", in.Name, in.Contexts[0].Name)

	path := configPath(t)
	var config *Config

	if _, err := os.Stat(path); os.IsNotExist(err) {
		config = &Config{
			CurrentContext: contextName,
			Project:        in.Project,
			Contact:        in.Contact,
			Environments:   []Environment{in},
		}
	} else if err != nil {
		return err
	} else {
		// existing config
		config, err := ReadConfig(t)
		if err != nil {
			return err
		}

		// overwrite current context
		config.CurrentContext = contextName

		// add environment
		config.Environments = append(config.Environments, in)
	}

	configBytes, err := yaml.Marshal(config)
	if err != nil {
		return err
	}

	err = utils.EnsureDirectory(t.ConfigPath(), 0700)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(path, configBytes, 0600)
}

func DefaultConfigSingleEnvSingleZoneAWSEUCentral() *Config {
	return &Config{
		CurrentContext: "devsingleeucentral-cluster",
		Project:        "tarmak-dev",
		Contact:        "tech@jetstack.io",
		Environments: []Environment{
			Environment{
				AWS: &AWSConfig{
					VaultPath: "jetstack/aws/jetstack-dev/sts/admin",
					Region:    "eu-central-1",
				},
				Name: "devsingleeucentral",
				Contexts: []Context{
					Context{
						Name:      "cluster",
						BaseImage: "centos-puppet-agent",
						Stacks: []Stack{
							Stack{
								State: &StackState{
									BucketPrefix: "jetstack-tarmak-",
									PublicZone:   "devsingleeucentral.dev.tarmak.org",
								},
							},
							Stack{
								Network: &StackNetwork{
									NetworkCIDR: "10.98.0.0/20",
									PrivateZone: "devsingleeucentral.dev.tarmak.local",
								},
							},
							Stack{
								Tools: &StackTools{},
							},
							Stack{
								Vault: &StackVault{},
							},
							Stack{
								Kubernetes: &StackKubernetes{},
							},
						},
					},
				},
			},
		},
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
					VaultPath:        "jetstack/aws/jetstack-dev/sts/admin",
					Region:           "eu-west-1",
					AvailabiltyZones: []string{"eu-west-1a", "eu-west-1b", "eu-west-1c"},
				},
				Name: "devsingle",
				Contexts: []Context{
					Context{
						Name:      "cluster",
						BaseImage: "centos-puppet-agent",
						Stacks: []Stack{
							Stack{
								State: &StackState{
									BucketPrefix: "jetstack-tarmak-",
									PublicZone:   "devsingle.dev.tarmak.org",
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
								Kubernetes: &StackKubernetes{},
								NodeGroups: DefaultKubernetesNodeGroupAWSOneMasterThreeEtcdThreeWorker(),
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
					VaultPath:        "jetstack/aws/jetstack-dev/sts/admin",
					Region:           "eu-west-1",
					AvailabiltyZones: []string{"eu-west-1a", "eu-west-1b", "eu-west-1c"},
				},
				Name: "devmulti",
				Contexts: []Context{
					Context{
						Name:      "hub",
						BaseImage: "centos-puppet-agent",
						Stacks: []Stack{
							Stack{
								State: &StackState{
									BucketPrefix: "jetstack-tarmak-",
									PublicZone:   "devmulti.dev.tarmak.org",
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
						Name:      "cluster",
						BaseImage: "centos-puppet-agent",
						Stacks: []Stack{
							Stack{
								Network: &StackNetwork{
									NetworkCIDR: "10.99.16.0/20",
									PeerContext: "hub",
								},
							},
							Stack{
								Kubernetes: &StackKubernetes{},
								NodeGroups: DefaultKubernetesNodeGroupAWSOneMasterThreeEtcdThreeWorker(),
							},
						},
					},
				},
			},
		},
	}
}

func DefaultKubernetesNodeGroupAWSOneMasterThreeEtcdThreeWorker() []NodeGroup {
	return []NodeGroup{
		NodeGroup{
			Count: 3,
			Role:  "etcd",
			AWS: &NodeGroupAWS{
				InstanceType: "m4.large",
				SpotPrice:    0.15,
			},
			Volumes: []Volume{
				Volume{
					AWS:  &VolumeAWS{Type: "gp2"},
					Name: "data",
					Size: 5,
				},
			},
		},
		NodeGroup{
			Count: 1,
			Role:  "master",
			AWS: &NodeGroupAWS{
				InstanceType: "m4.large",
				SpotPrice:    0.15,
			},
			Volumes: []Volume{
				Volume{
					AWS:  &VolumeAWS{Type: "gp2"},
					Name: "docker",
					Size: 10,
				},
			},
		},
		NodeGroup{
			Count: 3,
			Role:  "worker",
			AWS: &NodeGroupAWS{
				InstanceType: "m4.large",
				SpotPrice:    0.15,
			},
			Volumes: []Volume{
				Volume{
					AWS:  &VolumeAWS{Type: "gp2"},
					Name: "docker",
					Size: 50,
				},
			},
		},
		NodeGroup{
			Count: 1,
			Role:  "worker",
			Name:  "workert2",
			AWS: &NodeGroupAWS{
				InstanceType: "t2.large",
				SpotPrice:    0.15,
			},
			Volumes: []Volume{
				Volume{
					AWS:  &VolumeAWS{Type: "gp2"},
					Name: "docker",
					Size: 50,
				},
			},
		},
	}
}
