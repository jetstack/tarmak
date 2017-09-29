package interfaces

import (
	"io"
	"net"

	"github.com/Sirupsen/logrus"
	vault "github.com/hashicorp/vault/api"
	"github.com/jetstack-experimental/vault-unsealer/pkg/kv"

	clusterv1alpha1 "github.com/jetstack/tarmak/pkg/apis/cluster/v1alpha1"
	tarmakv1alpha1 "github.com/jetstack/tarmak/pkg/apis/tarmak/v1alpha1"
	"github.com/jetstack/tarmak/pkg/tarmak/role"
	"github.com/jetstack/tarmak/pkg/tarmak/utils/input"
)

type Cluster interface {
	Variables() map[string]interface{}
	Environment() Environment
	Name() string
	Validate() error
	Stacks() []Stack
	Stack(name string) Stack
	NetworkCIDR() *net.IPNet
	RemoteState(stackName string) string
	ConfigPath() string
	Config() *clusterv1alpha1.Cluster
	Images() []string // This returns all neccessary base images
	SSHConfigPath() string
	SSHHostKeysPath() string
	ClusterName() string
	Log() *logrus.Entry
	APITunnel() Tunnel
	Region() string
	Subnets() []clusterv1alpha1.Subnet // Return subnets per AZ
	Role(string) *role.Role
	Roles() []*role.Role
	InstancePools() []InstancePool
	ImageIDs() (map[string]string, error)
	Parameters() map[string]string
	ListHosts() ([]Host, error)
}

type Environment interface {
	Tarmak() Tarmak
	Location() string // this returns the location of the environment (e.g. the region)
	Variables() map[string]interface{}
	Provider() Provider
	Validate() error
	Name() string
	BucketPrefix() string
	Clusters() []Cluster
	Cluster(name string) (cluster Cluster, err error)
	SSHPrivateKeyPath() string
	SSHPrivateKey() (signer interface{})
	Log() *logrus.Entry
	StateStack() Stack
	VaultStack() Stack
	VaultRootToken() (string, error)
	Parameters() map[string]string
	VaultTunnel() (VaultTunnel, error)
	Config() *tarmakv1alpha1.Environment
	Type() string
}

type Provider interface {
	Cloud() string
	Name() string
	Parameters() map[string]string
	Region() string
	Validate() error
	RemoteStateBucketName() string
	RemoteStateBucketAvailable() (bool, error)
	RemoteState(namespace, clusterName, stackName string) string
	PublicZone() string
	Environment() ([]string, error)
	Variables() map[string]interface{}
	QueryImages(tags map[string]string) ([]tarmakv1alpha1.Image, error)
	VaultKV() (kv.Service, error)
	ListHosts() ([]Host, error)
	InstanceType(string) (string, error)
	VolumeType(string) (string, error)
	String() string
	AskEnvironmentLocation(Initialize) (string, error)
}

type Stack interface {
	Variables() map[string]interface{}
	Name() string
	Validate() error
	Cluster() Cluster
	RemoteState() string
	Log() *logrus.Entry
	VerifyPreDeploy() error
	VerifyPreDestroy() error
	VerifyPostDeploy() error
	VerifyPostDestroy() error
	SetOutput(map[string]interface{})
	Output() map[string]interface{}
	Roles() []*role.Role
	InstancePools() []InstancePool
}

type Tarmak interface {
	Variables() map[string]interface{}
	Log() *logrus.Entry
	RootPath() (string, error)
	ConfigPath() string
	Clusters() []Cluster
	Cluster() Cluster
	Environments() []Environment
	Environment() Environment
	Providers() []Provider
	Provider() Provider
	Terraform() Terraform
	Packer() Packer
	Puppet() Puppet
	Config() Config
	SSH() SSH
	HomeDirExpand(in string) (string, error)
	HomeDir() string
}

type Config interface {
	Cluster(environment string, name string) (cluster *clusterv1alpha1.Cluster, err error)
	Clusters(environment string) (clusters []*clusterv1alpha1.Cluster)
	AppendCluster(prov *clusterv1alpha1.Cluster) error
	UniqueClusterName(environment, name string) error
	Provider(name string) (provider *tarmakv1alpha1.Provider, err error)
	Providers() (providers []*tarmakv1alpha1.Provider)
	AppendProvider(prov *tarmakv1alpha1.Provider) error
	UniqueProviderName(name string) error
	ValidName(name, regex string) error
	ReadConfig() (*tarmakv1alpha1.Config, error)
	Environment(name string) (environment *tarmakv1alpha1.Environment, err error)
	Environments() (environments []*tarmakv1alpha1.Environment)
	AppendEnvironment(*tarmakv1alpha1.Environment) error
	UniqueEnvironmentName(name string) error
	CurrentClusterName() string
	CurrentEnvironmentName() string
	Contact() string
	Project() string
}

type Packer interface {
	IDs() (map[string]string, error)
	List() ([]tarmakv1alpha1.Image, error)
	Build() error
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
	Parameters() map[string]string
}

type Puppet interface {
	TarGz(io.Writer) error
}

type Kubectl interface {
}

type InstancePool interface {
	Config() *clusterv1alpha1.InstancePool
	TFName() string
	Name() string
	Image() string
	Role() *role.Role
	Volumes() []Volume
}

type Volume interface {
	Name() string
	Size() int
	Type() string
	Device() string
}

type Initialize interface {
	Input() *input.Input
	AskProjectName() (string, error)
	AskContact() (string, error)
	Config() Config
	Tarmak() Tarmak
	CurrentProvider() Provider
	CurrentEnvironment() Environment
}
