// Copyright Jetstack Ltd. See LICENSE for details.
package tarmak

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/go-multierror"
	"github.com/mitchellh/go-homedir"
	"github.com/sirupsen/logrus"

	tarmakv1alpha1 "github.com/jetstack/tarmak/pkg/apis/tarmak/v1alpha1"
	"github.com/jetstack/tarmak/pkg/packer"
	"github.com/jetstack/tarmak/pkg/puppet"
	"github.com/jetstack/tarmak/pkg/tarmak/assets"
	"github.com/jetstack/tarmak/pkg/tarmak/config"
	"github.com/jetstack/tarmak/pkg/tarmak/initialize"
	"github.com/jetstack/tarmak/pkg/tarmak/interfaces"
	"github.com/jetstack/tarmak/pkg/tarmak/kubectl"
	"github.com/jetstack/tarmak/pkg/tarmak/ssh"
	"github.com/jetstack/tarmak/pkg/terraform"
)

type Tarmak struct {
	homeDir         string
	rootPath        *string
	log             *logrus.Logger
	flags           *tarmakv1alpha1.Flags
	configDirectory string

	config    interfaces.Config
	terraform *terraform.Terraform
	puppet    *puppet.Puppet
	packer    *packer.Packer
	ssh       interfaces.SSH
	init      *initialize.Initialize
	kubectl   *kubectl.Kubectl

	environment interfaces.Environment
	cluster     interfaces.Cluster

	// function pointers for easier testing
	environmentByName func(string) (interfaces.Environment, error)
	providerByName    func(string) (interfaces.Provider, error)

	StopCh chan struct{}
}

var _ interfaces.Tarmak = &Tarmak{}

// allocate a new tarmak struct
func New(flags *tarmakv1alpha1.Flags) *Tarmak {
	t := &Tarmak{
		log:    logrus.New(),
		flags:  flags,
		StopCh: make(chan struct{}),
	}

	t.initializeModules()

	// set log level
	if flags.Verbose {
		t.log.SetLevel(logrus.DebugLevel)
	} else {
		t.log.SetLevel(logrus.InfoLevel)
	}

	// detect home directory
	homeDir, err := homedir.Dir()
	if err != nil {
		t.log.Fatal("unable to detect home directory: ", err)
	}
	t.homeDir = homeDir

	// set config directory

	// expand home directory
	t.configDirectory, err = homedir.Expand(flags.ConfigDirectory)
	if err != nil {
		t.log.Fatalf("unable to expand config directory ('%s'): %s", flags.ConfigDirectory, err)
	}

	// expand relative config path
	t.configDirectory, err = filepath.Abs(t.configDirectory)
	if err != nil {
		t.log.Fatalf("unable to expand relative config directory ('%s'): %s", t.configDirectory, err)
	}

	t.log.Level = logrus.DebugLevel
	t.log.Out = os.Stderr

	// read config, unless we are initialising the config
	t.config, err = config.New(t, flags)
	if err != nil {
		t.log.Fatal("unable to create tarmak: ", err)
	}

	// TODO: This needs to be validated
	_, err = t.config.ReadConfig()
	if err != nil {

		// TODO: This whole construct is really ugly, make this better soon
		if strings.Contains(err.Error(), "no such file or directory") {
			if flags.Initialize {
				return t
			}
			t.log.Fatal("unable to find an existing config, run 'tarmak init'")
		} else {
			t.log.Fatalf("failed to read config: %s", err)
		}

	}

	if flags.Initialize {
		return t
	}

	err = t.initializeConfig()
	if err != nil {
		t.log.Fatal("unable to initialize tarmak: ", err)
	}

	return t
}

// this initializes tarmak modules, they can be overridden in tests
func (t *Tarmak) initializeModules() {
	t.environmentByName = t.environmentByNameReal
	t.providerByName = t.providerByNameReal
	t.terraform = terraform.New(t, t.StopCh)
	t.packer = packer.New(t)
	t.ssh = ssh.New(t)
	t.puppet = puppet.New(t)
	t.kubectl = kubectl.New(t)
}

// Initialize default cluster, its environment and provider
func (t *Tarmak) initializeConfig() error {
	var err error

	// get current environment
	currentEnvironmentName, err := t.config.CurrentEnvironmentName()
	if err != nil {
		return fmt.Errorf("error retrieving current environment name: %s", err)
	}
	t.environment, err = t.EnvironmentByName(currentEnvironmentName)
	if err != nil {
		return err
	}

	clusterName, err := t.config.CurrentClusterName()
	if err != nil {
		return fmt.Errorf("failed to retrieve current cluster name: %s", err)
	}
	// init cluster
	t.cluster, err = t.environment.Cluster(clusterName)
	if err != nil {
		return fmt.Errorf("error finding current cluster '%s': %s", clusterName, err)
	}

	return nil
}

func (t *Tarmak) writeSSHConfigForClusterHosts() error {
	if err := t.ssh.WriteConfig(t.Cluster()); err != nil {
		clusterName, errCluster := t.config.CurrentClusterName()
		if errCluster != nil {
			return fmt.Errorf("failed to retrieve current cluster name: %s", errCluster)
		}
		return fmt.Errorf("failed to write ssh config for current cluster '%s': %v", clusterName, err)
	}
	return nil
}

// This initializes a new tarmak cluster
func (t *Tarmak) CmdClusterInit() error {
	i := initialize.New(t, os.Stdin, os.Stdout)
	t.init = i
	cluster, err := i.InitCluster()
	if err != nil {
		return err
	}

	t.log.Infof("successfully initialized cluster '%s'", cluster.ClusterName())

	err = t.config.SetCurrentCluster(cluster.ClusterName())
	if err != nil {
		return fmt.Errorf("error setting current cluster: %s", err)
	}
	return nil
}

func (t *Tarmak) CmdEnvironmentInit() error {
	i := initialize.New(t, os.Stdin, os.Stdout)
	environment, err := i.InitEnvironment()
	if err != nil {
		return err
	}

	t.log.Infof("successfully initialized environment '%s'", environment.Name())
	return nil
}

func (t *Tarmak) CmdProviderInit() error {
	i := initialize.New(t, os.Stdin, os.Stdout)
	provider, err := i.InitProvider()
	if err != nil {
		return err
	}

	t.log.Infof("successfully initialized provider '%s'", provider.Name())
	return nil
}

func (t *Tarmak) Puppet() interfaces.Puppet {
	return t.puppet
}

func (t *Tarmak) Config() interfaces.Config {
	return t.config
}

func (t *Tarmak) Packer() interfaces.Packer {
	return t.packer
}

func (t *Tarmak) Cluster() interfaces.Cluster {
	return t.cluster
}

func (t *Tarmak) Clusters() (clusters []interfaces.Cluster) {
	return clusters
}

// this builds a temporary directory with the needed assets that are built into the go binary
func (t *Tarmak) RootPath() (string, error) {
	if t.rootPath != nil {
		return *t.rootPath, nil
	}

	dir, err := ioutil.TempDir("", "tarmak-assets")
	if err != nil {
		return "", err
	}

	t.log.Debugf("created temporary directory: %s", dir)

	err = assets.RestoreAssets(dir, "")
	if err != nil {
		return "", err
	}

	t.log.Debugf("restored assets into directory: %s", dir)

	t.rootPath = &dir
	return *t.rootPath, nil
}

func (t *Tarmak) Log() *logrus.Entry {
	return t.log.WithField("app", "tarmak")
}

func (t *Tarmak) HomeDir() string {
	return t.homeDir
}

func (t *Tarmak) HomeDirExpand(in string) (string, error) {
	return homedir.Expand(in)
}

func (t *Tarmak) KeepContainers() bool {
	return t.flags.KeepContainers
}

func (t *Tarmak) ConfigPath() string {
	return t.configDirectory
}

func (t *Tarmak) Version() string {
	return t.flags.Version
}

func (t *Tarmak) Validate() error {
	var err error
	var result error

	err = t.Cluster().Validate()
	if err != nil {
		result = multierror.Append(result, err)
	}

	err = t.Cluster().Environment().Validate()
	if err != nil {
		result = multierror.Append(result, err)
	}

	return result
}

func (t *Tarmak) Verify() error {
	if err := t.Cluster().Environment().Verify(); err != nil {
		return fmt.Errorf("failed to verify tarmak provider: %s", err)
	}

	if err := t.Cluster().Verify(); err != nil {
		return fmt.Errorf("failed to verify tarmak cluster: %s", err)
	}

	if err := t.verifyImageExists(); err != nil {
		return err
	}

	return nil
}

func (t *Tarmak) Cleanup() {
	// clean up assets directory
	if t.rootPath != nil {
		if err := os.RemoveAll(*t.rootPath); err != nil {
			t.log.Warnf("error cleaning up assets directory: %s", err)
		}
		t.rootPath = nil
	}
}

func (t *Tarmak) Variables() map[string]interface{} {
	output := map[string]interface{}{}
	output["contact"] = t.config.Contact()
	output["project"] = t.config.Project()
	return output
}

func (t *Tarmak) Must(err error) {
	if err != nil {
		t.log.Fatal(err)
	}
}

func (t *Tarmak) CmdKubectl(args []string) error {
	if err := t.writeSSHConfigForClusterHosts(); err != nil {
		return err
	}
	return t.kubectl.Kubectl(args)
}
