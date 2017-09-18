package tarmak

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/hashicorp/go-multierror"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"

	"github.com/jetstack/tarmak/pkg/packer"
	"github.com/jetstack/tarmak/pkg/puppet"
	"github.com/jetstack/tarmak/pkg/tarmak/assets"
	"github.com/jetstack/tarmak/pkg/tarmak/config"
	"github.com/jetstack/tarmak/pkg/tarmak/environment"
	"github.com/jetstack/tarmak/pkg/tarmak/interfaces"
	"github.com/jetstack/tarmak/pkg/tarmak/kubectl"
	"github.com/jetstack/tarmak/pkg/tarmak/ssh"
	"github.com/jetstack/tarmak/pkg/terraform"
)

type Tarmak struct {
	homeDir  string
	rootPath *string
	log      *logrus.Logger

	config    *config.Config
	terraform *terraform.Terraform
	puppet    *puppet.Puppet
	packer    *packer.Packer
	ssh       interfaces.SSH
	cmd       *cobra.Command
	kubectl   *kubectl.Kubectl

	environment interfaces.Environment
	context     interfaces.Context
}

var _ interfaces.Tarmak = &Tarmak{}

func New(cmd *cobra.Command) *Tarmak {
	t := &Tarmak{
		log: logrus.New(),
		cmd: cmd,
	}

	// detect home directory
	homeDir, err := homedir.Dir()
	if err != nil {
		t.log.Fatal("unable to detect home directory: ", err)
	}
	t.homeDir = homeDir

	t.log.Level = logrus.DebugLevel
	t.log.Out = os.Stderr

	// TODO: enable me for init
	/*
		// return early for init
		if cmd.Name() == "init" {
			t.initialize = initialize.New(t)
			return t
		}
	*/

	// read config, unless we are initialising the config
	t.config, err = config.New(t)
	if err != nil {
		t.log.Fatal("unable to create tarmak: ", err)
	}

	// TODO: This needs to be validated
	_, err = t.config.ReadConfig()
	if err != nil {
		t.log.Fatal("unable to read config: ", err)
	}

	err = t.initialize()
	if err != nil {
		t.log.Fatal("unable to initialize tarmak: ", err)
	}

	t.terraform = terraform.New(t)
	t.packer = packer.New(t)
	t.ssh = ssh.New(t)
	t.puppet = puppet.New(t)
	t.kubectl = kubectl.New(t)

	return t
}

// Initialize default context, its environment and provider
func (t *Tarmak) initialize() error {
	var err error

	// get configs
	environmentName := t.config.CurrentEnvironmentName()
	environmentConfig, err := t.config.Environment(environmentName)
	if err != nil {
		return fmt.Errorf("error finding environment '%s'", environmentName)
	}

	contextConfigs := t.config.Contexts(environmentName)
	contextName := t.config.CurrentContextName()

	// init environment
	t.environment, err = environment.NewFromConfig(t, environmentConfig, contextConfigs)
	if err != nil {
		return fmt.Errorf("error initializing environment '%s': %s", environmentName, err)
	}

	// init context
	t.context, err = t.environment.Context(contextName)
	if err != nil {
		return fmt.Errorf("error finding current context '%s': %s", contextName, err)
	}

	return nil
}

// This initializes a new tarmak config
func (t *Tarmak) CmdInit() error {
	return fmt.Errorf("tarmak init needs refactoring")
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

func (t *Tarmak) Context() interfaces.Context {
	return t.context
}

func (t *Tarmak) Environment() interfaces.Environment {
	return t.environment
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

	// use same creation directory for all folders
	kubernetesEpoch := time.Unix(1437436800, 0)
	err = filepath.Walk(dir, func(path string, f os.FileInfo, err error) error {
		if f.IsDir() {
			err = os.Chtimes(path, kubernetesEpoch, kubernetesEpoch)
			if err != nil {
				return err
			}
		}
		return nil
	})
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

func (t *Tarmak) ConfigPath() string {
	return filepath.Join(t.HomeDir(), ".tarmak")
}

func (t *Tarmak) PackerBuild() {
	err := t.packer.Build()
	if err != nil {
		t.log.Fatalf("failed to query ami id: %s", err)
	}
}

func (t *Tarmak) Validate() error {
	var err error
	var result error

	err = t.Context().Validate()
	if err != nil {
		result = multierror.Append(result, err)
	}

	err = t.Context().Environment().Validate()
	if err != nil {
		result = multierror.Append(result, err)
	}

	return result
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
	return t.kubectl.Kubectl(args)
}
