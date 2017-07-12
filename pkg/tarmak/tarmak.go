package tarmak

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/hashicorp/go-multierror"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"

	"github.com/jetstack/tarmak/pkg/packer"
	"github.com/jetstack/tarmak/pkg/tarmak/config"
	"github.com/jetstack/tarmak/pkg/tarmak/environment"
	"github.com/jetstack/tarmak/pkg/tarmak/interfaces"
	"github.com/jetstack/tarmak/pkg/terraform"
)

type Tarmak struct {
	conf *config.Config

	homeDir   string
	rootPath  string
	log       *logrus.Logger
	terraform *terraform.Terraform
	packer    *packer.Packer
	cmd       *cobra.Command

	context      interfaces.Context
	environments []interfaces.Environment
}

var _ interfaces.Tarmak = &Tarmak{}

func New(cmd *cobra.Command) *Tarmak {
	t := &Tarmak{
		rootPath: "/home/christian/.golang/packages/src/github.com/jetstack/tarmak", // TODO: this should come from a go-bindata tree that is exported into tmp
		log:      logrus.New(),
		cmd:      cmd,
	}

	// detect home directory
	homeDir, err := homedir.Dir()
	if err != nil {
		t.log.Fatal("unabled to detect home directory: %s", err)
	}
	t.homeDir = homeDir

	t.log.Level = logrus.DebugLevel

	// TODO: read real config
	t.conf = config.DefaultConfigSingle()
	t.conf = config.DefaultConfigSingleEnvSingleZoneAWSEUCentral()

	// init environments
	for posEnvironment, _ := range t.conf.Environments {
		env, err := environment.NewFromConfig(t, &t.conf.Environments[posEnvironment])
		if err != nil {
			t.log.Fatal(err)
		}
		t.environments = append(t.environments, env)
	}

	// find context
	err = t.findContext()
	if err != nil {
		t.log.Fatal(err)
	}

	t.terraform = terraform.New(t)
	t.packer = packer.New(t)

	return t
}

func (t *Tarmak) Terraform() interfaces.Terraform {
	return t.packer
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

func (t *Tarmak) RootPath() string {
	return t.rootPath
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

func (t *Tarmak) TerraformApply(args []string) {
	if err := t.Validate(); err != nil {
		t.log.Fatal("could not validate config: ", err)
	}

	selectStacks, err := t.cmd.Flags().GetStringSlice("terraform-stacks")
	if err != nil {
		t.log.Fatal("could not find flag terraform-stacks: ", err)
	}

	t.discoverAMIID()
	stacks := t.Context().Stacks()
	for _, stack := range stacks {

		if len(selectStacks) > 0 {
			found := false
			for _, selectStack := range selectStacks {
				if selectStack == stack.Name() {
					found = true
				}
			}
			if !found {
				continue
			}
		}

		err := t.terraform.Apply(stack, args)
		if err != nil {
			t.log.Fatal(err)
		}
	}
}

func (t *Tarmak) TerraformDestroy(args []string) {
	selectStacks, err := t.cmd.Flags().GetStringSlice("terraform-stacks")
	if err != nil {
		t.log.Fatal("could not find flag terraform-stacks: ", err)
	}

	t.discoverAMIID()
	stacks := t.Context().Stacks()
	for posStack, _ := range stacks {
		stack := stacks[len(stacks)-posStack-1]
		if stack.Name() == config.StackNameState {
			t.log.Debugf("ignoring stack '%s'", stack.Name())
			continue
		}

		if len(selectStacks) > 0 {
			found := false
			for _, selectStack := range selectStacks {
				if selectStack == stack.Name() {
					found = true
				}
			}
			if !found {
				continue
			}
		}

		err := t.terraform.Destroy(stack, args)
		if err != nil {
			t.log.Fatal(err)
		}
		if err != nil {
			t.log.Fatal(err)
		}
	}
}

func (t *Tarmak) HomeDir() string {
	return t.homeDir
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
