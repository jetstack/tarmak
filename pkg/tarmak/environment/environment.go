package environment

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"path/filepath"

	"github.com/hashicorp/go-multierror"
	"github.com/mitchellh/go-homedir"

	"github.com/jetstack/tarmak/pkg/tarmak/config"
	"github.com/jetstack/tarmak/pkg/tarmak/context"
	"github.com/jetstack/tarmak/pkg/tarmak/interfaces"
	"github.com/jetstack/tarmak/pkg/tarmak/provider/aws"
	"github.com/jetstack/tarmak/pkg/tarmak/utils"
)

type Environment struct {
	conf *config.Environment

	contexts []interfaces.Context

	sshKeyPrivate *rsa.PrivateKey

	stackState interfaces.Stack
	stackVault interfaces.Stack
	stackTools interfaces.Stack

	provider interfaces.Provider

	tarmak interfaces.Tarmak
}

var _ interfaces.Environment = &Environment{}

func NewFromConfig(tarmak interfaces.Tarmak, conf *config.Environment) (*Environment, error) {
	e := &Environment{
		conf:   conf,
		tarmak: tarmak,
	}

	var result error

	networkCIDRs := []*net.IPNet{}

	for posContext, _ := range conf.Contexts {
		contextConf := &conf.Contexts[posContext]
		contextIntf, err := context.NewFromConfig(e, contextConf)

		if err != nil {
			result = multierror.Append(result, err)
			continue
		}
		e.contexts = append(e.contexts, contextIntf)

		networkCIDRs = append(networkCIDRs, contextIntf.NetworkCIDR())

		// loop through stacks
		for _, stack := range contextIntf.Stacks() {
			// ensure no multiple state stacks
			if stack.Name() == config.StackNameState {
				if e.stackState == nil {
					e.stackState = stack
				} else {
					result = multierror.Append(result, fmt.Errorf("environment '%s' has multiple state stacks", e.Name()))
				}
			}

			// ensure no multiple tools stacks
			if stack.Name() == config.StackNameTools {
				if e.stackTools == nil {
					e.stackTools = stack
				} else {
					result = multierror.Append(result, fmt.Errorf("environment '%s' has multiple tools stacks", e.Name()))
				}
			}

			// ensure no multiple vault stacks
			if stack.Name() == config.StackNameVault {
				if e.stackVault == nil {
					e.stackVault = stack
				} else {
					result = multierror.Append(result, fmt.Errorf("environment '%s' has multiple vault stacks", e.Name()))
				}
			}
		}
	}

	// ensure there is a state stack
	if e.stackState == nil {
		result = multierror.Append(result, fmt.Errorf("environment '%s' has no state stack", e.Name()))
	}

	// ensure there is a vault stack
	if e.stackTools == nil {
		result = multierror.Append(result, fmt.Errorf("environment '%s' has no tools stack", e.Name()))
	}

	// ensure there is a vault stack
	if e.stackVault == nil {
		result = multierror.Append(result, fmt.Errorf("environment '%s' has no vault stack", e.Name()))
	}

	// validate network overlap
	if err := utils.NetworkOverlap(networkCIDRs); err != nil {
		result = multierror.Append(result, err)
	}

	// init provider
	providers := []interfaces.Provider{}
	if conf.AWS != nil {
		provider, err := aws.NewFromConfig(e, conf.AWS)
		if err != nil {
			return nil, err
		}
		providers = append(providers, provider)
	}
	if conf.GCP != nil {
		return nil, errors.New("GCP not yet implemented :(")
	}

	if len(providers) < 1 {
		return nil, errors.New("please specify exactly one provider")
	}
	if len(providers) > 1 {
		return nil, fmt.Errorf("more than one provider given: %+v", providers)
	}
	e.provider = providers[0]

	return e, result

}

func (e *Environment) Name() string {
	return e.conf.Name
}

func (e *Environment) Provider() interfaces.Provider {
	return e.provider
}

func (e *Environment) Tarmak() interfaces.Tarmak {
	return e.tarmak
}

func (e *Environment) validateSSHKey() error {
	bytes, err := ioutil.ReadFile(e.conf.SSHKeyPath)
	if err != nil {
		return fmt.Errorf("unable to read ssh private key: %s", err)
	}

	block, _ := pem.Decode(bytes)
	if block == nil {
		return errors.New("failed to parse PEM block containing the ssh private key")
	}

	e.sshKeyPrivate, err = x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return fmt.Errorf("unable to parse private key: %s", err)
	}

	return fmt.Errorf("please implement me !!!")

}

func (e *Environment) Variables() map[string]interface{} {
	output := map[string]interface{}{}
	output["environment"] = e.Name()
	if e.conf.Contact != "" {
		output["contact"] = e.conf.Contact
	}
	if e.conf.Project != "" {
		output["project"] = e.conf.Project
	}

	for key, value := range e.provider.Variables() {
		output[key] = value
	}

	output["state_bucket"] = e.Provider().RemoteStateBucketName()
	output["state_context_name"] = e.stackState.Context().Name()
	output["tools_context_name"] = e.stackTools.Context().Name()
	output["vault_context_name"] = e.stackVault.Context().Name()
	return output
}

func (e *Environment) ConfigPath() string {
	return filepath.Join(e.tarmak.ConfigPath(), e.Name())
}

func (e *Environment) SSHPrivateKeyPath() string {
	if e.conf.SSHKeyPath == "" {
		return filepath.Join(e.ConfigPath(), "id_rsa")
	}

	dir, err := homedir.Expand(e.conf.SSHKeyPath)
	if err != nil {
		return e.conf.SSHKeyPath
	}
	return dir
}

func (e *Environment) Contexts() []interfaces.Context {
	return e.contexts
}

func (e *Environment) Validate() error {
	var result error

	err := e.Provider().Validate()
	if err != nil {
		result = multierror.Append(result, err)
	}

	return result
}

func (e *Environment) BucketPrefix() string {
	bucketPrefix, ok := e.stackState.Variables()["bucket_prefix"]
	if !ok {
		return ""
	}
	bucketPrefixString, ok := bucketPrefix.(string)
	if !ok {
		return ""
	}
	return bucketPrefixString
}
