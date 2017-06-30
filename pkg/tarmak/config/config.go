package config

import (
	"errors"
	"fmt"
	"net"

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
	CurrentContext string    `yaml:"currentContext,omitempty"` // <environment>-<name>
	Contexts       []Context `yaml:"contexts,omitempty"`
}

func (c *Config) GetContext() (*Context, error) {
	for _, context := range c.Contexts {
		if fmt.Sprintf("%s-%s", context.Environment, context.Name) == c.CurrentContext {
			return &context, nil
		}
	}
	return nil, fmt.Errorf("context with the name '%s' not found", c.CurrentContext)
}

func (c *Config) Validate() error {
	var result error

	// ensure no overlapping networks configured
	if err := c.validateNetworkOverlap(); err != nil {
		result = multierror.Append(result, err)
	}

	// ensure no required stack missing / duplicate stacks
	if err := c.validateRequiredStacks(); err != nil {
		result = multierror.Append(result, err)
	}

	return result
}

// check if a network overlaps
func (c *Config) validateNetworkOverlap() error {

	var result error

	netMap := map[string][]*net.IPNet{}

	for _, context := range c.Contexts {
		netCIDRs := []*net.IPNet{}
		for _, stack := range context.Stacks {
			if stack.Network == nil {
				continue
			}
			if stack.Network.NetworkCIDR == "" {
				continue
			}
			_, netCIDR, err := net.ParseCIDR(stack.Network.NetworkCIDR)
			if err != nil {
				result = multierror.Append(result, fmt.Errorf("failed parsing network CIDR '%s': %s", stack.Network.NetworkCIDR, err))
				continue
			}
			netCIDRs = append(netCIDRs, netCIDR)
		}

		// skip to the next when we don't have a network CIDR
		if len(netCIDRs) == 0 {
			continue
		}

		// append to network list in map
		if _, ok := netMap[context.Environment]; !ok {
			netMap[context.Environment] = netCIDRs
		} else {
			netMap[context.Environment] = append(netMap[context.Environment], netCIDRs...)
		}
	}

	// checking for overlaps
	for env, netCIDRs := range netMap {
		for i, _ := range netCIDRs {
			for j := i + 1; j < len(netCIDRs); j++ {
				// check for overlap per network
				if netCIDRs[i].Contains(netCIDRs[j].IP) || netCIDRs[j].Contains(netCIDRs[i].IP) {
					result = multierror.Append(result, fmt.Errorf(
						"network overlap in environment '%s', network '%s' overlaps with '%s'",
						env,
						netCIDRs[i].String(),
						netCIDRs[j].String(),
					))
				}
			}
		}
	}

	return result
}

func (c *Config) validateRequiredStacks() error {
	var result error
	stackStateMap := map[string]*StackState{}
	stackVaultMap := map[string]*StackVault{}

	// determine stacks per environment / context
	for posContext, _ := range c.Contexts {
		context := &c.Contexts[posContext]
		context.stackNetwork = nil
		for posStack, _ := range context.Stacks {
			stack := &context.Stacks[posStack]
			stackName, err := stack.StackName()
			if err != nil {
				result = multierror.Append(result, err)
				continue
			}
			if stackName == StackNameNetwork {
				if context.stackNetwork == nil {
					context.stackNetwork = stack.Network
				} else {
					result = multierror.Append(result, fmt.Errorf("context '%s-%s' has multiple network stacks", context.Environment, context.Name))
				}
			} else if stackName == StackNameState {
				_, ok := stackStateMap[context.Environment]
				if ok {
					result = multierror.Append(result, fmt.Errorf("context '%s-%s' overwrites an already existing state stack in the same environment", context.Environment, context.Name))
				} else {
					stackStateMap[context.Environment] = stack.State
				}
			} else if stackName == StackNameVault {
				_, ok := stackVaultMap[context.Environment]
				if ok {
					result = multierror.Append(result, fmt.Errorf("context '%s-%s' overwrites an already existing vault stack in the same environment", context.Environment, context.Name))
				} else {
					stackVaultMap[context.Environment] = stack.Vault
				}
			}
		}
	}

	// write back and check stacks per environment
	for posContext, _ := range c.Contexts {
		context := &c.Contexts[posContext]
		if context.stackNetwork == nil {
			result = multierror.Append(result, fmt.Errorf("context '%s-%s' has no network stack", context.Environment, context.Name))
		}
		if stack, ok := stackStateMap[context.Environment]; !ok {
			result = multierror.Append(result, fmt.Errorf("context '%s-%s' has no state stack in its environment '%s' ", context.Environment, context.Name, context.Environment))
		} else {
			context.stackState = stack
		}
		if stack, ok := stackVaultMap[context.Environment]; !ok {
			result = multierror.Append(result, fmt.Errorf("context '%s-%s' has no vault stack in its environment '%s' ", context.Environment, context.Name, context.Environment))
		} else {
			context.stackVault = stack
		}
	}

	return result

}

type Context struct {
	Name        string     `yaml:"name,omitempty"`        // only alphanumeric lowercase
	Environment string     `yaml:"environment,omitempty"` // only alphanumeric lowercase
	Stacks      []Stack    `yaml:"stacks,omitempty"`
	AWS         *AWSConfig `yaml:"aws,omitempty"`
	GCP         *GCPConfig `yaml:"gcp,omitempty"`

	stackNetwork *StackNetwork
	stackState   *StackState
	stackVault   *StackVault
}

func (c *Context) StateBucketPrefix() string {
	return c.stackState.BucketPrefix
}

func (c *Context) ProviderName() (string, error) {
	providers := []string{}
	if c.AWS != nil {
		providers = append(providers, ProviderNameAWS)
	}
	if c.GCP != nil {
		providers = append(providers, ProviderNameGCP)
	}

	if len(providers) < 1 {
		return "", errors.New("please specify exactly one provider")
	}
	if len(providers) > 1 {
		return "", fmt.Errorf("more than one provider given: %+v", providers)
	}

	return providers[0], nil
}

type GCPConfig struct {
	ProjectName string `yaml:"projectName,omitempty"`
	AccountName string `yaml:"accountName,omitempty"`
}

type Stack struct {
	State      *StackState      `yaml:"state,omitempty"`
	Network    *StackNetwork    `yaml:"network,omitempty"`
	Tools      *StackTools      `yaml:"tools,omitempty"`
	Vault      *StackVault      `yaml:"vault,omitempty"`
	Kubernetes *StackKubernetes `yaml:"kubernetes,omitempty"`
	Custom     *StackCustom     `yaml:"custom,omitempty"`
}

func (s *Stack) StackName() (string, error) {
	stacks := []string{}
	if s.State != nil {
		stacks = append(stacks, StackNameState)
	}
	if s.Network != nil {
		stacks = append(stacks, StackNameNetwork)
	}
	if s.Tools != nil {
		stacks = append(stacks, StackNameTools)
	}
	if s.Vault != nil {
		stacks = append(stacks, StackNameVault)
	}
	if s.Kubernetes != nil {
		stacks = append(stacks, StackNameKubernetes)
	}
	if s.Custom != nil {
		stacks = append(stacks, s.Custom.Name)
	}

	if len(stacks) < 1 {
		return "", errors.New("please exactly a single stack")
	}
	if len(stacks) > 1 {
		return "", fmt.Errorf("more than one stack given: %+v", stacks)
	}

	return stacks[0], nil

}

type StackTools struct {
}

type StackNetwork struct {
	PeerContext string `yaml:"peerContext,omitempty"`
	NetworkCIDR string `yaml:"networkCIDR,omitempty"`
	PrivateZone string `yaml:"privateZone,omitempty"`
}

type StackVault struct {
}

type StackState struct {
	BucketPrefix string `yaml:"bucketPrefix,omitempty"`
	PublicZone   string `yaml:"publicZone,omitempty"`
}

type StackKubernetes struct {
	EtcdCount     int     `yaml:"etcdCount,omitempty"`
	EtcdType      string  `yaml:"etcdType,omitempty"`
	EtcdSpotPrice float32 `yaml:"etcdSpotPrice,omitempty"`

	WorkerCount     int     `yaml:"workerCount,omitempty"`
	WorkerType      string  `yaml:"workerType,omitempty"`
	WorkerSpotPrice float32 `yaml:"workerSpotPrice,omitempty"`

	MasterCount     int     `yaml:"masterCount,omitempty"`
	MasterType      string  `yaml:"masterType,omitempty"`
	MasterSpotPrice float32 `yaml:"masterSpotPrice,omitempty"`
}

type StackCustom struct {
	Name string `yaml:"name,omitempty"`
	Path string `yaml:"path,omitempty"`
}

func DefaultConfigSingle() *Config {
	return &Config{
		CurrentContext: "devsingle-cluster",
		Contexts: []Context{
			Context{
				AWS: &AWSConfig{
					VaultPath: "jetstack/aws/jetstack-dev/sts/admin",
					Region:    "eu-west-1",
				},
				Environment: "devsingle",
				Name:        "cluster",
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
	}
}

func DefaultConfigHub() *Config {
	return &Config{
		CurrentContext: "devmulti-hub",
		Contexts: []Context{
			Context{
				AWS: &AWSConfig{
					VaultPath: "jetstack/aws/jetstack-dev/sts/admin",
					Region:    "eu-west-1",
				},
				Environment: "devmulti",
				Name:        "hub",
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
				AWS: &AWSConfig{
					VaultPath: "jetstack/aws/jetstack-dev/sts/admin",
					Region:    "eu-west-1",
				},
				Environment: "devmulti",
				Name:        "cluster",
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
	}
}
