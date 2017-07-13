package interfaces

import (
	"net"

	"github.com/Sirupsen/logrus"
)

type Context interface {
	Variables() map[string]interface{}
	Environment() Environment
	Name() string
	Validate() error
	Stacks() []Stack
	NetworkCIDR() *net.IPNet
	RemoteState(stackName string) string
	ConfigPath() string
	BaseImage() string
	SSHConfigPath() string
	SSHHostKeysPath() string
	SetImageID(string)
	ContextName() string
}

type Environment interface {
	Tarmak() Tarmak
	Variables() map[string]interface{}
	Provider() Provider
	Validate() error
	Name() string
	BucketPrefix() string
	Contexts() []Context
	SSHPrivateKeyPath() string
	SSHPrivateKey() (signer interface{})
}

type Provider interface {
	Name() string
	Region() string
	Validate() error
	RemoteStateBucketName() string
	RemoteStateBucketAvailable() (bool, error)
	RemoteState(contextName, stackName string) string
	Environment() ([]string, error)
	Variables() map[string]interface{}
	QueryImage(tags map[string]string) (string, error)
	ListHosts() ([]Host, error)
}

type Stack interface {
	Variables() map[string]interface{}
	Name() string
	Validate() error
	Context() Context
	RemoteState() string
}

type Tarmak interface {
	Variables() map[string]interface{}
	Log() *logrus.Entry
	RootPath() string
	ConfigPath() string
	Context() Context
	Environments() []Environment
	Terraform() Terraform
	Packer() Packer
}

type Packer interface {
}

type Terraform interface {
}

type Host interface {
	SSHConfig() string
}
