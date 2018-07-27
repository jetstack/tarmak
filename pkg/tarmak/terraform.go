// Copyright Jetstack Ltd. See LICENSE for details.
package tarmak

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/blang/semver"
	terraformVersion "github.com/hashicorp/terraform/version"
	"github.com/sirupsen/logrus"

	"github.com/jetstack/tarmak/pkg/tarmak/interfaces"
)

type CmdTerraform struct {
	StopCh chan struct{}
	log    *logrus.Entry
	tarmak *Tarmak
	args   []string
	ctx    interfaces.CancellationContext
}

func (t *Tarmak) Terraform() interfaces.Terraform {
	return t.terraform
}

func (t *Tarmak) NewCmdTerraform(args []string) *CmdTerraform {
	return &CmdTerraform{
		tarmak: t,
		log:    t.Log(),
		args:   args,
		ctx:    t.Context(),
	}

}

func (c *CmdTerraform) Plan() error {
	c.log.Info("validate steps")
	if err := c.tarmak.Validate(); err != nil {
		return fmt.Errorf("failed to validate tarmak: %s", err)
	}

	select {
	case <-c.ctx.Done():
		return c.ctx.Err()
	default:
	}

	c.log.Info("verify steps")
	if err := c.tarmak.Verify(); err != nil {
		return err
	}

	select {
	case <-c.ctx.Done():
		return c.ctx.Err()
	default:
	}

	c.log.Info("write SSH config")
	if err := c.tarmak.writeSSHConfigForClusterHosts(); err != nil {
		return err
	}

	c.log.Info("running plan")
	err := c.tarmak.terraform.Plan(c.tarmak.Cluster())
	if err != nil {
		return err
	}

	return nil
}

func (c *CmdTerraform) Apply() error {
	c.log.Info("validate steps")
	if err := c.tarmak.Validate(); err != nil {
		return fmt.Errorf("failed to validate tarmak: %s", err)
	}
	select {
	case <-c.ctx.Done():
		return c.ctx.Err()
	default:
	}

	c.log.Info("verify steps")
	if err := c.tarmak.Verify(); err != nil {
		return err
	}

	select {
	case <-c.ctx.Done():
		return c.ctx.Err()
	default:
	}

	c.log.Info("write SSH config")
	if err := c.tarmak.writeSSHConfigForClusterHosts(); err != nil {
		return err
	}

	c.log.Info("running apply")
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

	select {
	case <-c.ctx.Done():
		return c.ctx.Err()
	default:
	}

	// wait for convergance in every mode
	err := c.tarmak.Cluster().WaitForConvergance()
	if err != nil {
		return err
	}

	return nil
}

func (c *CmdTerraform) Destroy() error {
	c.log.Info("validate steps")
	if err := c.tarmak.Validate(); err != nil {
		return fmt.Errorf("failed to validate tarmak: %s", err)
	}

	select {
	case <-c.ctx.Done():
		return c.ctx.Err()
	default:
	}

	c.log.Info("verify steps")
	if err := c.tarmak.Verify(); err != nil {
		return err
	}

	select {
	case <-c.ctx.Done():
		return c.ctx.Err()
	default:
	}

	c.log.Info("write SSH config")
	if err := c.tarmak.writeSSHConfigForClusterHosts(); err != nil {
		return err
	}

	c.log.Info("running destroy")

	c.tarmak.cluster.Log().Info("running destroy")
	err := c.tarmak.terraform.Destroy(c.tarmak.Cluster())
	if err != nil {
		return err
	}

	return nil
}

func (c *CmdTerraform) Shell(args []string) error {
	if err := c.verifyTerraformBinaryVersion(); err != nil {
		return err
	}

	if err := c.tarmak.writeSSHConfigForClusterHosts(); err != nil {
		return err
	}

	err := c.tarmak.terraform.Shell(c.tarmak.Cluster())
	if err != nil {
		return err
	}

	return nil
}

func (c *CmdTerraform) verifyTerraformBinaryVersion() error {

	cmd := exec.Command("terraform", "version")
	cmd.Env = os.Environ()
	cmdOutput := &bytes.Buffer{}
	cmd.Stdout = cmdOutput

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to run 'terraform version': %s. Please make sure that Terraform is installed", err)
	}

	reader := bufio.NewReader(cmdOutput)
	versionLine, _, err := reader.ReadLine()
	if err != nil {
		return fmt.Errorf("failed to read 'terraform version' output: %s", err)
	}

	terraformBinaryVersion := strings.TrimPrefix(string(versionLine), "Terraform v")
	terraformVendoredVersion := terraformVersion.Version

	terraformBinaryVersionSemver, err := semver.Make(terraformBinaryVersion)
	if err != nil {
		return fmt.Errorf("failed to parse Terraform binary version: %s", err)
	}
	terraformVendoredVersionSemver, err := semver.Make(terraformVendoredVersion)
	if err != nil {
		return fmt.Errorf("failed to parse Terraform vendored version: %s", err)
	}

	// we need binary version == vendored version
	if terraformBinaryVersionSemver.GT(terraformVendoredVersionSemver) {
		return fmt.Errorf("Terraform binary version (%s) is greater than vendored version (%s). Please downgrade binary version to %s", terraformBinaryVersion, terraformVendoredVersion, terraformVendoredVersion)
	} else if terraformBinaryVersionSemver.LT(terraformVendoredVersionSemver) {
		return fmt.Errorf("Terraform binary version (%s) is less than vendored version (%s). Please upgrade binary version to %s", terraformBinaryVersion, terraformVendoredVersion, terraformVendoredVersion)
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
