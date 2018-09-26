// Copyright Jetstack Ltd. See LICENSE for details.
package tarmak

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/blang/semver"
	terraformVersion "github.com/hashicorp/terraform/version"
	"github.com/sirupsen/logrus"

	"github.com/jetstack/tarmak/pkg/tarmak/interfaces"
	"github.com/jetstack/tarmak/pkg/tarmak/utils/input"
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
		ctx:    t.CancellationContext(),
	}
}

func (c *CmdTerraform) Plan() (returnCode int, err error) {
	if err := c.setup(); err != nil {
		return 1, err
	}

	c.log.Info("running plan")
	changesNeeded, err := c.tarmak.terraform.Plan(c.tarmak.Cluster())
	if changesNeeded {
		return 2, err
	} else {
		return 0, err
	}
}

func (c *CmdTerraform) Apply() error {
	if err := c.setup(); err != nil {
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
	if err := c.setup(); err != nil {
		return err
	}

	c.log.Info("running destroy")
	err := c.tarmak.terraform.Destroy(c.tarmak.Cluster())
	if err != nil {
		return err
	}

	return nil
}

func (c *CmdTerraform) Shell() error {
	if err := c.setup(); err != nil {
		c.log.Warnf("error setting up tarmak for terrafrom shell: %v", err)
	}

	if err := c.verifyTerraformBinaryVersion(); err != nil {
		return err
	}

	err := c.tarmak.terraform.Shell(c.tarmak.Cluster())
	if err != nil {
		return err
	}

	return nil
}

func (c *CmdTerraform) ForceUnlock() error {
	if err := c.setup(); err != nil {
		return err
	}

	if len(c.args) != 1 {
		return fmt.Errorf("expected single lock ID argument, got=%d", len(c.args))
	}

	in := input.New(os.Stdin, os.Stdout)
	query := fmt.Sprintf(`Attempting force-unlock using lock ID [%s]
Are you sure you want to force-unlock the remote state? This can be potentially dangerous!`, c.args[0])
	doUnlock, err := in.AskYesNo(&input.AskYesNo{
		Default: false,
		Query:   query,
	})
	if err != nil {
		return err
	}

	if !doUnlock {
		c.log.Infof("aborting force unlock")
		return nil
	}

	c.tarmak.cluster.Log().Info("running force-unlock")
	err = c.tarmak.terraform.ForceUnlock(c.tarmak.Cluster(), c.args[0])
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

func (c *CmdTerraform) setup() error {
	type step struct {
		log string
		f   func() error
	}

	for _, s := range []step{
		{"validating tarmak config", c.tarmak.Validate},
		{"verifying tarmak config", c.tarmak.Verify},
		{"writing SSH config", c.tarmak.writeSSHConfigForClusterHosts},
		{"ensuring remote resources", c.tarmak.EnsureRemoteResources},
	} {
		c.log.Info(s.log)
		if err := s.f(); err != nil {
			return err
		}

		select {
		case <-c.ctx.Done():
			return c.ctx.Err()
		default:
		}
	}

	return nil
}
