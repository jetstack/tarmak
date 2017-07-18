package interfaces

import (
	"net"

	"github.com/Sirupsen/logrus"
	"github.com/jetstack-experimental/vault-unsealer/pkg/kv"
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
	Log() *logrus.Entry
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
	Log() *logrus.Entry
	StateStack() Stack
	VaultStack() Stack
	VaultRootToken() (string, error)
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
	VaultKV() (kv.Service, error)
	ListHosts() ([]Host, error)
}

type Stack interface {
	Variables() map[string]interface{}
	Name() string
	Validate() error
	Context() Context
	RemoteState() string
	Log() *logrus.Entry
	VerifyPost() error
	SetOutput(map[string]interface{})
	Output() map[string]interface{}
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
	SSH() SSH
	HomeDirExpand(in string) (string, error)
	HomeDir() string
}

type Packer interface {
}

type Terraform interface {
	Output(stack Stack) (map[string]interface{}, error)
}

type SSH interface {
	WriteConfig() error
	PassThrough([]string)
	Tunnel(hostname string, destination string, destinationPort int) Tunnel
	Execute(host string, cmd string, args []string) (returnCode int, err error)
}

type Tunnel interface {
	Start() error
	Stop() error
	Port() int
}

type Host interface {
	ID() string
	Hostname() string
	User() string
	Roles() []string
	SSHConfig() string
}
