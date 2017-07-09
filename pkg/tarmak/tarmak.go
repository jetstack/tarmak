package tarmak

import (
	"github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/jetstack/tarmak/pkg/packer"
	"github.com/jetstack/tarmak/pkg/tarmak/config"
	"github.com/jetstack/tarmak/pkg/terraform"
)

type Tarmak struct {
	rootPath  string
	log       *logrus.Logger
	context   *config.Context
	terraform *terraform.Terraform
	packer    *packer.Packer
	cmd       *cobra.Command
}

var _ config.Tarmak = &Tarmak{}

func New(cmd *cobra.Command) *Tarmak {

	log := logrus.New()
	log.Level = logrus.DebugLevel

	// TODO: read real config
	myConfig := config.DefaultConfigSingle()
	err := myConfig.Validate()
	if err != nil {
		log.Fatal(err)
	}

	context, err := myConfig.GetContext()
	if err != nil {
		log.Fatal(err)
	}

	t := &Tarmak{
		rootPath: "/home/christian/.golang/packages/src/github.com/jetstack/tarmak", // TODO: this should come from a go-bindata tree that is exported into tmp
		log:      log,
		context:  context,
		cmd:      cmd,
	}
	t.terraform = terraform.New(t)
	t.packer = packer.New(t)
	return t
}

func (t *Tarmak) Terraform() config.Terraform {
	return t.packer
}

func (t *Tarmak) Packer() config.Packer {
	return t.packer
}

func (t *Tarmak) Context() *config.Context {
	return t.context
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
	selectStacks, err := t.cmd.Flags().GetStringSlice("terraform-stacks")
	if err != nil {
		t.log.Fatal("could not find flag terraform-stacks: ", err)
	}

	t.discoverAMIID()
	for posStack, _ := range t.context.Stacks {
		stack := &t.context.Stacks[posStack]

		if len(selectStacks) > 0 {
			found := false
			for _, selectStack := range selectStacks {
				if selectStack == stack.StackName() {
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
	for posStack, _ := range t.context.Stacks {
		stack := &t.context.Stacks[len(t.context.Stacks)-posStack-1]
		if stack.StackName() == config.StackNameState {
			t.log.Debugf("ignoring stack '%s'", stack.StackName())
			continue
		}

		if len(selectStacks) > 0 {
			found := false
			for _, selectStack := range selectStacks {
				if selectStack == stack.StackName() {
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
