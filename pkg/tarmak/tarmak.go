package tarmak

import (
	"github.com/Sirupsen/logrus"

	"github.com/jetstack/tarmak/pkg/packer"
	"github.com/jetstack/tarmak/pkg/tarmak/config"
	"github.com/jetstack/tarmak/pkg/terraform"
)

type Tarmak struct {
	rootPath string
	log      *logrus.Logger
	context  *config.Context
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

	return &Tarmak{
		rootPath: "/home/christian/.golang/packages/src/github.com/jetstack/tarmak", // TODO: this should come from a go-bindata tree that is exported into tmp
		log:      log,
		context:  context,
	}
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

func (t *Tarmak) TerraformApply() {
	tf := terraform.New(t)
	for posStack, _ := range t.context.Stacks {
		err := tf.Apply(&t.context.Stacks[posStack])
		if err != nil {
			t.log.Fatal(err)
		}
	}
}

func (t *Tarmak) TerraformDestroy() {
	tf := terraform.New(t)
	stacks := t.context.Stacks[0:2]
	for posStack, _ := range stacks {
		err := tf.Destroy(&stacks[len(stacks)-posStack-1])
		if err != nil {
			t.log.Fatal(err)
		}
	}
}

func (t *Tarmak) PackerBuild() {
	p := packer.New(t)
	_, err := p.Build()
	//_, err := p.Build()
	if err != nil {
		t.log.Fatalf("failed to query ami id: %s", err)
	}
}

func (t *Tarmak) PackerQuery() {
	p := packer.New(t)
	_, err := p.QueryAMIID()
	if err != nil {
		t.log.Fatalf("failed to query ami id: %s", err)
	}
}
