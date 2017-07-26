package tarmak

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/hashicorp/go-multierror"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"

	"github.com/jetstack/tarmak/pkg/packer"
	"github.com/jetstack/tarmak/pkg/puppet"
	"github.com/jetstack/tarmak/pkg/tarmak/assets"
	"github.com/jetstack/tarmak/pkg/tarmak/config"
	"github.com/jetstack/tarmak/pkg/tarmak/environment"
	"github.com/jetstack/tarmak/pkg/tarmak/initialize"
	"github.com/jetstack/tarmak/pkg/tarmak/interfaces"
	"github.com/jetstack/tarmak/pkg/tarmak/kubectl"
	"github.com/jetstack/tarmak/pkg/tarmak/ssh"
	"github.com/jetstack/tarmak/pkg/terraform"
)

type Tarmak struct {
	conf *config.Config

	homeDir    string
	rootPath   *string
	log        *logrus.Logger
	terraform  *terraform.Terraform
	puppet     *puppet.Puppet
	packer     *packer.Packer
	ssh        interfaces.SSH
	cmd        *cobra.Command
	kubectl    *kubectl.Kubectl
	initialize *initialize.Init

	context      interfaces.Context
	environments []interfaces.Environment
}

var _ interfaces.Tarmak = &Tarmak{}

func (t *Tarmak) MergeEnvironment(inInterface interface{}) error {
	switch in := inInterface.(type) {
	case config.Environment:
		return config.MergeEnvironment(t, in)
	}

	return fmt.Errorf("unexpected type %T for parameter", inInterface)
}

func (t *Tarmak) Init() error {
	return t.initialize.Run()
}

func New(cmd *cobra.Command) *Tarmak {
	t := &Tarmak{
		log: logrus.New(),
		cmd: cmd,
	}

	// detect home directory
	homeDir, err := homedir.Dir()
	if err != nil {
		t.log.Fatal("unabled to detect home directory: ", err)
	}
	t.homeDir = homeDir

	t.log.Level = logrus.DebugLevel
	t.log.Out = os.Stderr

	// return early for init
	if cmd.Name() == "init" {
		t.initialize = initialize.New(t)
		return t
	}

	// read config, unless we are initialising the config
	conf, err := config.ReadConfig(t)
	if err != nil {
		t.log.Fatal("unabled to read config: ", err)
	}

	if err := t.initFromConfig(conf); err != nil {
		t.log.Fatal("unabled to validate config: ", err)
	}

	t.terraform = terraform.New(t)
	t.packer = packer.New(t)
	t.ssh = ssh.New(t)
	t.puppet = puppet.New(t)
	t.initialize = initialize.New(t)
	t.kubectl = kubectl.New(t)

	return t
}

func (t *Tarmak) initFromConfig(cfg *config.Config) error {
	var result error

	// init environments
	for posEnvironment, _ := range cfg.Environments {
		env, err := environment.NewFromConfig(t, &cfg.Environments[posEnvironment])
		if err != nil {
			result = multierror.Append(result, err)
		}
		t.environments = append(t.environments, env)
	}
	if result != nil {
		return result
	}
	t.conf = cfg

	// find context
	if err := t.findContext(); err != nil {
		result = multierror.Append(result, err)
	}
	return result
}

func (t *Tarmak) Puppet() interfaces.Puppet {
	return t.puppet
}

func (t *Tarmak) Packer() interfaces.Packer {
	return t.packer
}

func (t *Tarmak) Context() interfaces.Context {
	return t.context
}

func (t *Tarmak) findContext() error {
	for _, environment := range t.environments {
		if !strings.HasPrefix(t.conf.CurrentContext, fmt.Sprintf("%s-", environment.Name())) {
			continue
		}
		for _, context := range environment.Contexts() {
			if context.ContextName() == t.conf.CurrentContext {
				t.context = context
				return nil
			}
		}
	}
	return fmt.Errorf("context '%s' not found", t.conf.CurrentContext)
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

func (t *Tarmak) discoverAMIID() {
	amiID, err := t.packer.QueryAMIID()
	if err != nil {
		t.log.Fatal("could not find a matching ami: ", err)
	}
	t.Context().SetImageID(amiID)
}

func (t *Tarmak) HomeDir() string {
	return t.homeDir
}

func (t *Tarmak) HomeDirExpand(in string) (string, error) {
	return homedir.Expand(in)
}

func (t *Tarmak) Environments() []interfaces.Environment {
	return t.environments
}

func (t *Tarmak) ConfigPath() string {
	return filepath.Join(t.HomeDir(), ".tarmak")
}

func (t *Tarmak) PackerBuild() {
	_, err := t.packer.Build()
	if err != nil {
		t.log.Fatalf("failed to query ami id: %s", err)
	}
}

func (t *Tarmak) PackerQuery() {
	_, err := t.packer.QueryAMIID()
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
	if t.conf.Contact != "" {
		output["contact"] = t.conf.Contact
	}
	if t.conf.Project != "" {
		output["project"] = t.conf.Project
	}
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
