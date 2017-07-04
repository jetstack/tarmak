package tarmak

import (
	"github.com/Sirupsen/logrus"

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
}

var _ config.Tarmak = &Tarmak{}

func New() *Tarmak {

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

func (t *Tarmak) TerraformApply() {
	t.discoverAMIID()
	for posStack, _ := range t.context.Stacks {
		err := t.terraform.Apply(&t.context.Stacks[posStack])
		if err != nil {
			t.log.Fatal(err)
		}
	}
}

func (t *Tarmak) TerraformDestroy() {
	t.discoverAMIID()
	stacks := t.context.Stacks[0:2]
	for posStack, _ := range stacks {
		err := t.terraform.Destroy(&stacks[len(stacks)-posStack-1])
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
