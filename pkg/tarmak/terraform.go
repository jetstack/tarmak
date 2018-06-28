// Copyright Jetstack Ltd. See LICENSE for details.
package tarmak

import (
	"errors"
	"fmt"

	"github.com/jetstack/tarmak/pkg/tarmak/interfaces"
)

type CmdTerraform struct {
	StopCh chan struct{}
	tarmak *Tarmak
	args   []string
}

func (t *Tarmak) Terraform() interfaces.Terraform {
	return t.terraform
}

func (t *Tarmak) NewCmdTerraform(args []string) *CmdTerraform {
	return &CmdTerraform{
		StopCh: t.StopCh,
		tarmak: t,
		args:   args,
	}

}

func (c *CmdTerraform) Plan() error {
	if err := c.tarmak.writeSSHConfigForClusterHosts(); err != nil {
		return err
	}

	if err := c.tarmak.verifyImageExists(); err != nil {
		return err
	}

	if err := c.tarmak.Validate(); err != nil {
		return fmt.Errorf("failed to validate tarmak: %s", err)
	}

	if err := c.tarmak.Cluster().Verify(); err != nil {
		return fmt.Errorf("failed to validate tarmak cluster: %s", err)
	}

	c.tarmak.cluster.Log().Info("running plan")
	err := c.tarmak.terraform.Plan(c.tarmak.Cluster())
	if err != nil {
		return err
	}

	return nil
}

func (c *CmdTerraform) Apply() error {
	if err := c.tarmak.writeSSHConfigForClusterHosts(); err != nil {
		return err
	}

	if err := c.tarmak.verifyImageExists(); err != nil {
		return err
	}

	if err := c.tarmak.Validate(); err != nil {
		return fmt.Errorf("failed to validate tarmak: %s", err)
	}

	if err := c.tarmak.Cluster().Verify(); err != nil {
		return fmt.Errorf("failed to validate tarmak cluster: %s", err)
	}

	c.tarmak.cluster.Log().Info("running apply")
	// run terraform apply always, do not run it when in configuration only mode
	if !c.tarmak.flags.Cluster.Apply.ConfigurationOnly {
		err := c.tarmak.terraform.Apply(c.tarmak.Cluster())
		if err != nil {
			return err
		}
	}

	// upload tar gz only if terraform hasn't uploaded it yet
	if c.tarmak.flags.Cluster.Apply.ConfigurationOnly {
		err := c.tarmak.Cluster().UploadConfiguration()
		if err != nil {
			return err
		}
	}

	// reapply config expect if we are in infrastructure only
	if !c.tarmak.flags.Cluster.Apply.InfrastructureOnly {
		err := c.tarmak.Cluster().ReapplyConfiguration()
		if err != nil {
			return err
		}
	}

	// wait for convergance in every mode
	err := c.tarmak.Cluster().WaitForConvergance()
	if err != nil {
		return err
	}

	return nil
}

func (c *CmdTerraform) Destroy() error {
	if err := c.tarmak.writeSSHConfigForClusterHosts(); err != nil {
		return err
	}

	if err := c.tarmak.Validate(); err != nil {
		return fmt.Errorf("failed to validate tarmak: %s", err)
	}

	if err := c.tarmak.Cluster().Verify(); err != nil {
		return fmt.Errorf("failed to validate tarmak cluster: %s", err)
	}

	c.tarmak.cluster.Log().Info("running destroy")

	err := c.tarmak.terraform.Destroy(c.tarmak.Cluster())
	if err != nil {
		return err
	}
	return nil
}

func (c *CmdTerraform) Shell(args []string) error {
	if err := c.tarmak.writeSSHConfigForClusterHosts(); err != nil {
		return err
	}

	err := c.tarmak.terraform.Shell(c.tarmak.Cluster())
	if err != nil {
		return err
	}
	return nil
}

func (t *Tarmak) verifyImageExists() error {
	images, err := t.Packer().List()
	if err != nil {
		return err
	}

	if len(images) == 0 {
		return errors.New("no images found")
	}

	return nil
}
