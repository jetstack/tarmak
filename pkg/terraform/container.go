package terraform

import (
	"fmt"
	"os"
	"path/filepath"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/docker/pkg/archive"

	tarmakDocker "github.com/jetstack/tarmak/pkg/docker"
	"github.com/jetstack/tarmak/pkg/tarmak/config"
)

type TerraformContainer struct {
	*tarmakDocker.AppContainer
	t     *Terraform
	stack *config.Stack
	log   *log.Entry
}

func (tc *TerraformContainer) Plan(destroy bool) (changesNeeded bool, err error) {

	args := []string{"plan", "-out=terraform.plan", "-detailed-exitcode", "-input=false"}

	if destroy {
		args = append(args, "-destroy")
	}

	// adds parameters as CLI args
	for key, value := range tc.stack.TerraformVars(tc.t.tarmak.Context().TerraformVars()) {
		switch v := value.(type) {
		case map[string]string:
			val := "{"
			for mkey, mval := range v {
				val += fmt.Sprintf(" %s = \"%s\",", mkey, mval)
			}
			val = val[:len(val)-1]
			val += "}"
			args = append(args, "-var", fmt.Sprintf("%s=%s", key, val))
		case string:
			args = append(args, "-var", fmt.Sprintf("%s=%s", key, v))
		default:
			tc.log.Warnf("ignoring unknown var type %t", v)
		}
	}

	returnCode, err := tc.Execute("terraform", args)
	if err != nil {
		return false, err
	}

	if returnCode == 0 {
		return false, nil
	}
	if returnCode == 2 {
		return true, nil
	}
	return false, fmt.Errorf("unexpected return code: exp=0/2, act=%d", returnCode)
}

func (tc *TerraformContainer) Apply() error {
	returnCode, err := tc.Execute("terraform", []string{"apply", "-input=false", "terraform.plan"})
	if err != nil {
		return err
	}
	if exp, act := 0, returnCode; exp != act {
		return fmt.Errorf("unexpected return code: exp=%d, act=%d", exp, act)
	}
	return nil
}

func (tc *TerraformContainer) Init() error {
	returnCode, err := tc.Execute("terraform", []string{"init", "-input=false"})
	if err != nil {
		return err
	}
	if exp, act := 0, returnCode; exp != act {
		return fmt.Errorf("unexpected return code: exp=%d, act=%d", exp, act)
	}
	return nil
}

func (tc *TerraformContainer) InitForceCopy() error {
	returnCode, err := tc.Execute("terraform", []string{"init", "-force-copy", "-input=false"})
	if err != nil {
		return err
	}
	if exp, act := 0, returnCode; exp != act {
		return fmt.Errorf("unexpected return code: exp=%d, act=%d", exp, act)
	}
	return nil
}

func (tc *TerraformContainer) CopyRemoteState(content string) error {
	remoteStateTar, err := tarmakDocker.TarStreamFromFile("terraform_remote_state.tf", content)
	if err != nil {
		return err
	}

	err = tc.UploadToContainer(remoteStateTar, "/terraform")
	if err != nil {
		return err
	}
	tc.log.Debug("copied remote state config into container")

	return nil
}

func (tc *TerraformContainer) prepare() error {
	// get aws secrets
	if environmentProvider, err := tc.t.tarmak.Context().ProviderEnvironment(); err != nil {
		return fmt.Errorf("error getting environment secrets from provider: %s", err)
	} else {
		tc.Env = append(tc.Env, environmentProvider...)
	}
	tc.log.WithField("environment", tc.Env).Debug("")

	// set default commandpfals
	tc.Cmd = []string{"sleep", "3600"}
	tc.WorkingDir = "/terraform"

	// build terraform image if needed
	tc.log.Debug("prepare container")

	err := tc.AppContainer.Prepare()
	if err != nil {
		return err
	}

	// tar terraform manifests
	tarOpts := &archive.TarOptions{
		Compression:  archive.Uncompressed,
		NoLchown:     true,
		IncludeFiles: []string{"."},
	}

	terraformDir := filepath.Clean(filepath.Join(tc.t.tarmak.RootPath(), "terraform/aws-centos", tc.stack.StackName()))
	tc.log = tc.log.WithField("terraform-dir", terraformDir)

	terraformDirInfo, err := os.Stat(terraformDir)
	if err != nil {
		return err
	}
	if !terraformDirInfo.IsDir() {
		return fmt.Errorf("path '%s' is not a directory", terraformDir)
	}

	terraformTar, err := archive.TarWithOptions(terraformDir, tarOpts)
	if err != nil {
		return err
	}

	err = tc.UploadToContainer(terraformTar, "/terraform")
	if err != nil {
		return err
	}
	tc.log.Debug("copied terraform manifests into container")

	err = tc.Start()
	if err != nil {
		return fmt.Errorf("error starting container: %s", err)
	}

	return nil
}
