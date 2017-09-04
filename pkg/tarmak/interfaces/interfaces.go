package interfaces

import (
	"io"
	"net"

	"github.com/Sirupsen/logrus"
	vault "github.com/hashicorp/vault/api"
	"github.com/jetstack-experimental/vault-unsealer/pkg/kv"
	"github.com/jetstack/tarmak/pkg/tarmak/role"

	clusterv1alpha1 "github.com/jetstack/tarmak/pkg/apis/cluster/v1alpha1"
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
	APITunnel() Tunnel
	Region() string
	Subnets() []clusterv1alpha1.Subnet
}

type Environment interface {
	Tarmak() Tarmak
	Variables() map[string]interface{}
	Provider() Provider
	Validate() error
	Name() string
	BucketPrefix() string
	Contexts() []Context
	CurrentContext() Context
	SSHPrivateKeyPath() string
	SSHPrivateKey() (signer interface{})
	Log() *logrus.Entry
	StateStack() Stack
	VaultStack() Stack
	VaultRootToken() (string, error)
	VaultTunnel() (VaultTunnel, error)
}

type Provider interface {
	Name() string
	Region() string
	Validate() error
	RemoteStateBucketName() string
	RemoteStateBucketAvailable() (bool, error)
	RemoteState(namespace, clusterName, stackName string) string
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
	VerifyPreDeploy() error
	VerifyPreDestroy() error
	VerifyPostDeploy() error
	VerifyPostDestroy() error
	SetOutput(map[string]interface{})
	Output() map[string]interface{}
	Role(string) *role.Role
	Roles() []*role.Role
	NodeGroups() []NodeGroup
}

type Tarmak interface {
	Variables() map[string]interface{}
	Log() *logrus.Entry
	RootPath() (string, error)
	ConfigPath() string
	Context() Context
	Environments() []Environment
	Terraform() Terraform
	Packer() Packer
	Puppet() Puppet
	SSH() SSH
	HomeDirExpand(in string) (string, error)
	HomeDir() string
	MergeEnvironment(interface{}) error
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
	BindAddress() string
}

type VaultTunnel interface {
	Tunnel
	VaultClient() *vault.Client
}

type Host interface {
	ID() string
	Hostname() string
	User() string
	Roles() []string
	SSHConfig() string
}

type Puppet interface {
	TarGz(io.Writer) error
}

type Kubectl interface {
}

type NodeGroup interface {
	Name() string
	Role() *role.Role
	Volumes() []Volume
}

type Volume interface {
	Name() string
	Size() int
	Type() string
	Device() string
}
