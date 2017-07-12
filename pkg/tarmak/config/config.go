package config

import ()

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
					KeyName:   "jetstack_nonprod",
				},
				SSHKeyPath: "~/.ssh/id_jetstack_nonprod",
				Name:       "devsingleeucentral",
				Contexts: []Context{
					Context{
						Name:      "cluster",
						BaseImage: "centos-puppet-agent",
						Stacks: []Stack{
							Stack{
								State: &StackState{
									BucketPrefix: "jetstack-tarmak-",
									PublicZone:   "devsingleeucentral.dev.tarmak.io",
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
					KeyName:          "jetstack_nonprod",
				},
				SSHKeyPath: "~/.ssh/id_jetstack_nonprod",
				Name:       "devsingle",
				Contexts: []Context{
					Context{
						Name:      "cluster",
						BaseImage: "centos-puppet-agent",
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
					VaultPath:        "jetstack/aws/jetstack-dev/sts/admin",
					Region:           "eu-west-1",
					AvailabiltyZones: []string{"eu-west-1a", "eu-west-1b", "eu-west-1c"},
					KeyName:          "jetstack_nonprod",
				},
				SSHKeyPath: "~/.ssh/id_jetstack_nonprod",
				Name:       "devmulti",
				Contexts: []Context{
					Context{
						Name:      "hub",
						BaseImage: "centos-puppet-agent",
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
