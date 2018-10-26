// Copyright Jetstack Ltd. See LICENSE for details.
package tarmak

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/blang/semver"
	terraformVersion "github.com/hashicorp/terraform/version"
	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"

	"github.com/jetstack/tarmak/pkg/tarmak/interfaces"
	"github.com/jetstack/tarmak/pkg/tarmak/utils"
	"github.com/jetstack/tarmak/pkg/tarmak/utils/consts"
	"github.com/jetstack/tarmak/pkg/tarmak/utils/input"
)

type CmdTarmak struct {
	*Tarmak

	log    *logrus.Entry
	args   []string
	pflags *pflag.FlagSet
	ctx    interfaces.CancellationContext
}

func (t *Tarmak) NewCmdTarmak(pflags *pflag.FlagSet, args []string) *CmdTarmak {
	return &CmdTarmak{
		Tarmak: t,
		log:    t.Log(),
		args:   args,
		pflags: pflags,
		ctx:    t.CancellationContext(),
	}
}

func (c *CmdTarmak) Plan() (returnCode int, err error) {
	if err := c.setup(); err != nil {
		return 1, err
	}

	changesNeeded, err := c.terraform.Plan(c.Cluster(), false)
	if changesNeeded {
		return 2, err
	} else {
		return 0, err
	}
}

func (c *CmdTarmak) Apply() error {
	err := c.setup()
	if err != nil {
		return err
	}

	// assume a change so that we wait for convergence in configuration only
	hasChanged := true
	// run terraform apply always, do not run it when in configuration only mode
	if !c.flags.Cluster.Apply.ConfigurationOnly {
		hasChanged, err = c.terraform.Apply(c.Cluster())
		if err != nil {
			return err
		}
	}

	// upload tar gz only if terraform hasn't uploaded it yet
	if c.flags.Cluster.Apply.ConfigurationOnly {
		err := c.Cluster().UploadConfiguration()
		if err != nil {
			return err
		}
	}

	// reapply config expect if we are in infrastructure only
	if !c.flags.Cluster.Apply.InfrastructureOnly {
		err := c.Cluster().ReapplyConfiguration()
		if err != nil {
			return err
		}
	}

	select {
	case <-c.ctx.Done():
		return c.ctx.Err()
	default:
	}

	// wait for convergance if flag enabled and has changed
	if hasChanged && c.flags.Cluster.Apply.WaitForConvergence {
		err := c.Cluster().WaitForConvergance()
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *CmdTarmak) Destroy() error {
	if err := c.setup(); err != nil {
		return err
	}

	if err := c.terraform.Destroy(c.Cluster()); err != nil {
		return err
	}

	return nil
}

func (c *CmdTarmak) Shell() error {
	if err := c.setup(); err != nil {
		c.log.Warnf("error setting up tarmak for terrafrom shell: %v", err)
	}

	if err := c.verifyTerraformBinaryVersion(); err != nil {
		return err
	}

	err := c.terraform.Shell(c.Cluster())
	if err != nil {
		return err
	}

	return nil
}

func (c *CmdTarmak) ForceUnlock() error {
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

	err = c.terraform.ForceUnlock(c.Cluster(), c.args[0])
	if err != nil {
		return err
	}

	return nil
}

func (c *CmdTarmak) ImagesBuild() error {
	requiredImages := c.cluster.Images()
	c.args = utils.RemoveDuplicateStrings(c.args)

	// rebuild existing flag so build the de-duplicated list of existing and
	// given args
	if c.flags.Cluster.Images.Build.RebuildExisting {
		return c.packer.Build(
			utils.RemoveDuplicateStrings(
				append(requiredImages, c.args...),
			))
	}

	images, err := c.Packer().List()
	if err != nil {
		return err
	}

	var currentImages []string
	for _, i := range images {
		if c.cluster.AmazonEBSEncrypted() == i.Encrypted {
			currentImages = append(currentImages, i.BaseImage)
		}
	}

	var missingImages []string
	for _, i := range requiredImages {
		if !utils.SliceContains(currentImages, i) {
			missingImages = append(missingImages, i)
		}
	}

	if len(c.args) == 0 {
		if len(missingImages) == 0 {
			c.log.Infof("all images have been built for this cluster")
			return nil
		}

		return c.packer.Build(missingImages)
	}

	var alreadyBuilt []string
	for _, i := range c.args {
		if utils.SliceContains(currentImages, i) {
			alreadyBuilt = append(alreadyBuilt, i)
		}
	}

	if len(alreadyBuilt) != 0 {
		in := input.New(os.Stdin, os.Stdout)
		query := fmt.Sprintf(`The following images have already been built %s
Are you sure you want to re-build them?`, alreadyBuilt)
		b, err := in.AskYesNo(&input.AskYesNo{
			Default: false,
			Query:   query,
		})
		if err != nil {
			return err
		}

		if !b {
			c.log.Info("aborting building images")
			return nil
		}
	}

	return c.packer.Build(c.args)
}

func (c *CmdTarmak) ImagesDestroy() error {
	return c.Provider().DestroyImages(c.args)
}

func (c *CmdTarmak) Kubectl() error {
	if err := c.writeSSHConfigForClusterHosts(); err != nil {
		return err
	}

	return c.kubectl.Kubectl(c.args, c.kubePublicAPIEndpoint())
}

func (c *CmdTarmak) Kubeconfig() error {
	var err error

	path := c.flags.Cluster.Kubeconfig.Path
	if path == consts.DefaultKubeconfigPath {
		path = c.kubectl.ConfigPath()

	} else {
		path, err = filepath.Abs(path)
		if err != nil {
			return fmt.Errorf("failed to get absolute path of custom path: %s", err)
		}

		c.log.Debugf("using custom kubeconfig path %s", path)
	}

	kubeconfig, err := c.kubectl.Kubeconfig(path, c.kubePublicAPIEndpoint())
	if err != nil {
		return err
	}

	fmt.Printf("%s\n", kubeconfig)

	return nil
}

func (c *CmdTarmak) kubePublicAPIEndpoint() bool {
	// first set bool to what we have set in the config
	publicEndpoint := false
	if k := c.Cluster().Config().Kubernetes; k != nil && k.APIServer != nil {
		publicEndpoint = k.APIServer.Public
	}

	// if the flag default is different to the config AND we have changed the
	// flag (overridden), we set the bool and warn we are using a different
	// setting than the config
	if p := c.flags.PublicAPIEndpoint; publicEndpoint != p &&
		c.pflags.Changed(consts.KubeconfigFlagName) {
		c.log.Warnf("overriding %s from tarmak config to %v", consts.KubeconfigFlagName, p)
		publicEndpoint = p
	}

	return publicEndpoint
}

func (c *CmdTarmak) verifyTerraformBinaryVersion() error {
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

func (c *CmdTarmak) setup() error {
	type step struct {
		log string
		f   func() error
	}

	for _, s := range []step{
		{"validating tarmak config", c.Validate},
		{"verifying tarmak config", c.Verify},
		{"writing SSH config", c.writeSSHConfigForClusterHosts},
		{"ensuring remote resources", c.EnsureRemoteResources},
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
