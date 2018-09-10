// Copyright Jetstack Ltd. See LICENSE for details.
package tarmak

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/blang/semver"
	"github.com/hashicorp/go-multierror"
	terraformVersion "github.com/hashicorp/terraform/version"

	"github.com/jetstack/tarmak/pkg/tarmak/interfaces"
)

func (t *Tarmak) Terraform() interfaces.Terraform {
	return t.terraform
}

func (t *Tarmak) CmdTerraformPlan(args []string, ctx context.Context) error {
	t.cluster.Log().Info("validate steps")
	if err := t.Validate(); err != nil {
		return fmt.Errorf("failed to validate tarmak: %s", err)
	}

	t.cluster.Log().Info("verify steps")
	if err := t.Verify(); err != nil {
		return err
	}

	t.cluster.Log().Info("write SSH config")
	if err := t.writeSSHConfigForClusterHosts(); err != nil {
		return err
	}

	t.cluster.Log().Info("running plan")
	err := t.terraform.Plan(t.Cluster())
	if err != nil {
		return err
	}

	return nil
}

func (t *Tarmak) CmdTerraformApply(args []string, ctx context.Context) error {
	t.cluster.Log().Info("validate steps")
	if err := t.Validate(); err != nil {
		return fmt.Errorf("failed to validate tarmak: %s", err)
	}

	t.cluster.Log().Info("verify steps")
	if err := t.Verify(); err != nil {
		return err
	}

	if t.flags.Cluster.Apply.SpotPricing {
		t.cluster.Log().Info("calculating instance pool spot prices")

		var result *multierror.Error
		for _, i := range t.Cluster().InstancePools() {
			if err := i.CalculateSpotPrice(); err != nil {
				result = multierror.Append(result, err)
			}
		}

		if result != nil {
			return result.ErrorOrNil()
		}
	}

	t.cluster.Log().Info("write SSH config")
	if err := t.writeSSHConfigForClusterHosts(); err != nil {
		return err
	}

	t.cluster.Log().Info("running apply")
	// run terraform apply always, do not run it when in configuration only mode
	if !t.flags.Cluster.Apply.ConfigurationOnly {
		err := t.terraform.Apply(t.Cluster())
		if err != nil {
			return err
		}
	}

	// upload tar gz only if terraform hasn't uploaded it yet
	if t.flags.Cluster.Apply.ConfigurationOnly {
		err := t.Cluster().UploadConfiguration()
		if err != nil {
			return err
		}
	}

	// reapply config expect if we are in infrastructure only
	if !t.flags.Cluster.Apply.InfrastructureOnly {
		err := t.Cluster().ReapplyConfiguration()
		if err != nil {
			return err
		}
	}

	// wait for convergance in every mode
	err := t.Cluster().WaitForConvergance()
	if err != nil {
		return err
	}

	return nil
}

func (t *Tarmak) CmdTerraformDestroy(args []string, ctx context.Context) error {
	t.cluster.Log().Info("validate steps")
	if err := t.Validate(); err != nil {
		return fmt.Errorf("failed to validate tarmak: %s", err)
	}

	t.cluster.Log().Info("verify steps")
	if err := t.Verify(); err != nil {
		return err
	}

	t.cluster.Log().Info("write SSH config")
	if err := t.writeSSHConfigForClusterHosts(); err != nil {
		return err
	}

	t.cluster.Log().Info("running destroy")

	err := t.terraform.Destroy(t.Cluster())
	if err != nil {
		return err
	}
	return nil
}

func (t *Tarmak) CmdTerraformShell(args []string) error {

	if err := t.verifyTerraformBinaryVersion(); err != nil {
		return err
	}

	if err := t.writeSSHConfigForClusterHosts(); err != nil {
		return err
	}

	err := t.terraform.Shell(t.Cluster())
	if err != nil {
		return err
	}

	return nil
}

func (t *Tarmak) verifyTerraformBinaryVersion() error {

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
