package context

import (
	"errors"
	"fmt"
	"net"
	"path/filepath"

	"github.com/Sirupsen/logrus"
	"github.com/hashicorp/go-multierror"

	"github.com/jetstack/tarmak/pkg/tarmak/config"
	"github.com/jetstack/tarmak/pkg/tarmak/interfaces"
	"github.com/jetstack/tarmak/pkg/tarmak/stack"
)

type Context struct {
	conf *config.Context

	stacks []interfaces.Stack

	stackNetwork interfaces.Stack
	environment  interfaces.Environment
	imageID      *string
	networkCIDR  *net.IPNet
	log          *logrus.Entry
}

var _ interfaces.Context = &Context{}

func NewFromConfig(environment interfaces.Environment, conf *config.Context) (*Context, error) {
	context := &Context{
		conf:        conf,
		environment: environment,
		log:         environment.Log().WithField("context", conf.Name),
	}

	var result error

	for posStack, _ := range conf.Stacks {
		stackConf := &conf.Stacks[posStack]
		stackIntf, err := stack.NewFromConfig(context, stackConf)
		if err != nil {
			result = multierror.Append(result, err)
			continue
		}
		context.stacks = append(context.stacks, stackIntf)

		if stackIntf.Name() == config.StackNameNetwork {
			if context.stackNetwork == nil {
				context.stackNetwork = stackIntf
			} else {
				result = multierror.Append(result, fmt.Errorf("context '%s' has multiple network stacks", context.Name()))
			}
		}
	}

	if context.stackNetwork == nil {
		result = multierror.Append(result, fmt.Errorf("context '%s' has no network stack", context.Name()))
	} else {
		_, err := context.getNetworkCIDR()
		if err != nil {
			result = multierror.Append(result, fmt.Errorf("context '%s' has an incorrect network CIDR: %s", context.Name(), err))
		}

	}

	return context, nil
}

func (c *Context) RemoteState(stackName string) string {
	return c.Environment().Provider().RemoteState(c.Name(), stackName)
}

func (c *Context) BaseImage() string {
	return c.conf.BaseImage
}

func (c *Context) getNetworkCIDR() (*net.IPNet, error) {
	if c.stackNetwork == nil {
		return nil, errors.New("no network stack found")
	}

	netIntf, ok := c.stackNetwork.Variables()["network"]
	if !ok {
		return nil, errors.New("no network variable in stack network found")
	}

	net, ok := netIntf.(*net.IPNet)
	if !ok {
		return nil, errors.New("network variable has unexpected typ")
	}

	return net, nil
}

func (c *Context) NetworkCIDR() *net.IPNet {
	return c.networkCIDR
}

func (c *Context) SetImageID(imageID string) {
	c.imageID = &imageID
}

func (c *Context) Validate() error {
	return nil
}

func (c *Context) Stacks() []interfaces.Stack {
	return c.stacks
}

func (c *Context) Environment() interfaces.Environment {
	return c.environment
}

func (c *Context) ContextName() string {
	return fmt.Sprintf("%s-%s", c.environment.Name(), c.conf.Name)
}

func (c *Context) Name() string {
	return c.conf.Name
}

func (c *Context) ConfigPath() string {
	return filepath.Join(c.Environment().Tarmak().ConfigPath(), c.ContextName())
}

func (c *Context) SSHConfigPath() string {
	return filepath.Join(c.ConfigPath(), "ssh_config")
}

func (c *Context) SSHHostKeysPath() string {
	return filepath.Join(c.ConfigPath(), "ssh_known_hosts")
}

func (c *Context) Log() *logrus.Entry {
	return c.log
}

func (c *Context) Variables() map[string]interface{} {
	output := c.environment.Variables()

	if c.conf.Contact != "" {
		output["contact"] = c.conf.Contact
	}
	if c.conf.Project != "" {
		output["project"] = c.conf.Project
	}

	if c.imageID != nil {
		output["centos_ami"] = map[string]string{
			c.environment.Provider().Region(): *c.imageID,
		}
	}

	output["name"] = c.Name()

	return output
}
