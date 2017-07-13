package terraform

import (
	"fmt"

	"github.com/Sirupsen/logrus"

	tarmakDocker "github.com/jetstack/tarmak/pkg/docker"
	"github.com/jetstack/tarmak/pkg/tarmak/config"
	"github.com/jetstack/tarmak/pkg/tarmak/interfaces"
)

type Terraform struct {
	*tarmakDocker.App
	log    *logrus.Entry
	tarmak interfaces.Tarmak
}

func New(tarmak interfaces.Tarmak) *Terraform {
	log := tarmak.Log().WithField("module", "terraform")

	app := tarmakDocker.NewApp(
		tarmak,
		log,
		"jetstack/tarmak-terraform",
		"terraform",
	)

	return &Terraform{
		App:    app,
		log:    log,
		tarmak: tarmak,
	}
}

func (t *Terraform) NewContainer(stack interfaces.Stack) *TerraformContainer {
	c := &TerraformContainer{
		AppContainer: t.Container(),
		t:            t,
		log:          t.log.WithField("stack", stack.Name()),
		stack:        stack,
	}
	c.AppContainer.SetLog(t.log.WithField("stack", stack.Name()))
	return c
}

func (t *Terraform) Apply(stack interfaces.Stack, args []string) error {
	return t.planApply(stack, args, false)
}

func (t *Terraform) Destroy(stack interfaces.Stack, args []string) error {
	return t.planApply(stack, args, true)
}

func (t *Terraform) planApply(stack interfaces.Stack, args []string, destroy bool) error {
	c := t.NewContainer(stack)

	if err := c.prepare(); err != nil {
		return fmt.Errorf("error preparing container: %s", err)
	}
	defer c.CleanUpSilent(t.log)

	initialStateStack := false
	// check for initial state run on first deployment
	if !destroy && stack.Name() == config.StackNameState {
		remoteStateAvail, err := t.tarmak.Context().Environment().Provider().RemoteStateBucketAvailable()
		if err != nil {
			return fmt.Errorf("error finding remote state: %s", err)
		}
		if !remoteStateAvail {
			initialStateStack = true
			c.log.Infof("running state stack for the first time, by passing remote state")
		}
	}

	if !initialStateStack {
		err := c.CopyRemoteState(stack.RemoteState())

		if err != nil {
			return fmt.Errorf("error while copying remote state: %s", err)
		}
		c.log.Debug("copied remote state into container")
	}

	if err := c.Init(); err != nil {
		return fmt.Errorf("error while terraform init: %s", err)
	}

	// check for destroying the state stack
	if destroy && stack.Name() == config.StackNameState {
		c.log.Infof("moving remote state to local")

		err := c.CopyRemoteState("")
		if err != nil {
			return fmt.Errorf("error while copying empty remote state: %s", err)
		}
		c.log.Debug("copied empty remote state into container")

		if err := c.InitForceCopy(); err != nil {
			return fmt.Errorf("error while terraform init -force-copy: %s", err)
		}
	}

	changesNeeded, err := c.Plan(args, destroy)
	if err != nil {
		return fmt.Errorf("error while terraform plan: %s", err)
	}

	if changesNeeded {
		if err := c.Apply(); err != nil {
			return fmt.Errorf("error while terraform apply: %s", err)
		}
	}

	// upload state if it was an inital state run
	if initialStateStack {
		err := c.CopyRemoteState(stack.RemoteState())
		if err != nil {
			return fmt.Errorf("error while copying remote state: %s", err)
		}
		c.log.Debug("copied remote state into container")

		if err := c.InitForceCopy(); err != nil {
			return fmt.Errorf("error while terraform init -force-copy: %s", err)
		}
	}

	// verify that state has been run successfully
	if err := stack.VerifyPost(); err != nil {
		return fmt.Errorf("verify of stack %s failed: %s", stack.Name(), err)
	}

	return nil
}
