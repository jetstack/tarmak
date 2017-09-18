package terraform

import (
	"fmt"

	"github.com/Sirupsen/logrus"

	tarmakv1alpha1 "github.com/jetstack/tarmak/pkg/apis/tarmak/v1alpha1"
	tarmakDocker "github.com/jetstack/tarmak/pkg/docker"
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

func (t *Terraform) Output(stack interfaces.Stack) (map[string]interface{}, error) {
	if output := stack.Output(); output != nil {
		return output, nil
	}

	c := t.NewContainer(stack)

	if err := c.prepare(); err != nil {
		return nil, fmt.Errorf("error preparing container: %s", err)
	}
	defer c.CleanUpSilent(t.log)

	err := c.CopyRemoteState(stack.RemoteState())
	if err != nil {
		return nil, fmt.Errorf("error while copying remote state: %s", err)
	}
	c.log.Debug("copied remote state into container")

	if err := c.Init(); err != nil {
		return nil, fmt.Errorf("error while terraform init: %s", err)
	}

	output, err := c.Output()
	stack.SetOutput(output)
	if err != nil {
		return nil, fmt.Errorf("error while getting terraform output: %s", err)
	}

	return output, nil

}

func (t *Terraform) Shell(stack interfaces.Stack, args []string) error {
	c := t.NewContainer(stack)

	if err := c.prepare(); err != nil {
		return fmt.Errorf("error preparing container: %s", err)
	}
	defer c.CleanUpSilent(t.log)

	if err := c.CopyRemoteState(stack.RemoteState()); err != nil {
		return fmt.Errorf("error while copying remote state: %s", err)
	}
	c.log.Debug("copied remote state into container")

	if err := c.Init(); err != nil {
		return fmt.Errorf("error while terraform init: %s", err)
	}

	return c.Shell()
}

func (t *Terraform) planApply(stack interfaces.Stack, args []string, destroy bool) error {
	c := t.NewContainer(stack)

	if destroy {
		if err := stack.VerifyPreDestroy(); err != nil {
			return fmt.Errorf("verify of stack %s failed: %s", stack.Name(), err)
		}
	} else {
		if err := stack.VerifyPreDeploy(); err != nil {
			return fmt.Errorf("verify of stack %s failed: %s", stack.Name(), err)
		}
	}

	if err := c.prepare(); err != nil {
		return fmt.Errorf("error preparing container: %s", err)
	}
	defer c.CleanUpSilent(t.log)

	err := c.CopyRemoteState(stack.RemoteState())

	if err != nil {
		return fmt.Errorf("error while copying remote state: %s", err)
	}
	c.log.Debug("copied remote state into container")

	if err := c.Init(); err != nil {
		return fmt.Errorf("error while terraform init: %s", err)
	}

	// check for destroying the state stack
	if destroy && stack.Name() == tarmakv1alpha1.StackNameState {
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

	// verify that state has been run successfully
	if !destroy {
		output, err := c.Output()
		stack.SetOutput(output)
		if err != nil {
			return fmt.Errorf("error while getting terraform output: %s", err)
		}
		t.log.WithFields(output).Debug("terraform output")

		if err := stack.VerifyPostDeploy(); err != nil {
			return fmt.Errorf("verify of stack %s failed: %s", stack.Name(), err)
		}
	} else {
		if err := stack.VerifyPostDestroy(); err != nil {
			return fmt.Errorf("verify of stack %s failed: %s", stack.Name(), err)
		}

	}

	return nil
}
