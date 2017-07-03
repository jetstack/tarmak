package terraform

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/docker/pkg/archive"
	"github.com/fsouza/go-dockerclient"

	tarmakDocker "github.com/jetstack/tarmak/pkg/docker"
	"github.com/jetstack/tarmak/pkg/tarmak/config"
)

type TerraformContainer struct {
	t               *Terraform
	dockerContainer *docker.Container
	stack           *config.Stack
	log             *log.Entry
}

func (tc *TerraformContainer) CleanUp() error {
	return fmt.Errorf("unimplemented: %s", "CleanUp")
}

func (tc *TerraformContainer) Plan(destroy bool) (changesNeeded bool, err error) {

	args := []string{"plan", "-out=terraform.plan", "-detailed-exitcode", "-input=false"}

	if destroy {
		args = append(args, "-destroy")
	} else {
		// adds parameters as CLI args
		for key, value := range tc.stack.TerraformVars(tc.t.context.TerraformVars()) {
			switch v := value.(type) {
			case string:
				args = append(args, "-var", fmt.Sprintf("%s=%s", key, v))
			default:
				tc.log.Warnf("ignoring unknown var type %t", v)
			}
		}
	}

	returnCode, err := tc.execCmd("terraform", args)
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
	returnCode, err := tc.execCmd("terraform", []string{"apply", "-input=false", "terraform.plan"})
	if err != nil {
		return err
	}
	if exp, act := 0, returnCode; exp != act {
		return fmt.Errorf("unexpected return code: exp=%d, act=%d", exp, act)
	}
	return nil
}

func (tc *TerraformContainer) Init() error {
	returnCode, err := tc.execCmd("terraform", []string{"init", "-input=false"})
	if err != nil {
		return err
	}
	if exp, act := 0, returnCode; exp != act {
		return fmt.Errorf("unexpected return code: exp=%d, act=%d", exp, act)
	}
	return nil
}

func (tc *TerraformContainer) InitForceCopy() error {
	returnCode, err := tc.execCmd("terraform", []string{"init", "-force-copy", "-input=false"})
	if err != nil {
		return err
	}
	if exp, act := 0, returnCode; exp != act {
		return fmt.Errorf("unexpected return code: exp=%d, act=%d", exp, act)
	}
	return nil
}

func (tc *TerraformContainer) execCmd(cmd string, args []string) (returnCode int, err error) {
	command := []string{cmd}
	command = append(command, args...)
	tc.log.WithField("command", command).Debug()
	exec, err := tc.t.dockerClient.CreateExec(docker.CreateExecOptions{
		Cmd:          command,
		AttachStdin:  false,
		AttachStdout: true,
		AttachStderr: true,
		Container:    tc.dockerContainer.ID,
	})
	if err != nil {
		return -1, err
	}

	stdoutReader, stdoutWriter := io.Pipe()
	stdoutScanner := bufio.NewScanner(stdoutReader)
	go func() {
		for stdoutScanner.Scan() {
			var action string
			if cmd == "terraform" {
				action = fmt.Sprintf("%s-%s", cmd, args[0])
			} else {
				action = cmd
			}
			tc.log.WithField("action", action).Debug(stdoutScanner.Text())
		}
	}()

	err = tc.t.dockerClient.StartExec(exec.ID, docker.StartExecOptions{
		ErrorStream:  stdoutWriter,
		OutputStream: stdoutWriter,
	})
	stdoutReader.Close()
	stdoutWriter.Close()

	if err != nil {
		return -1, fmt.Errorf("error starting exec: %s", err)
	}

	execInspect, err := tc.t.dockerClient.InspectExec(exec.ID)
	if err != nil {
		return -1, fmt.Errorf("error inspecting exec: %s", err)
	}

	return execInspect.ExitCode, nil
}

func (tc *TerraformContainer) CopyRemoteState(content string) error {
	remoteStateTar, err := tarmakDocker.TarStreamFromFile("terraform_remote_state.tf", content)
	if err != nil {
		return err
	}

	err = tc.t.dockerClient.UploadToContainer(tc.dockerContainer.ID, docker.UploadToContainerOptions{
		InputStream: remoteStateTar,
		Path:        "/terraform",
	})
	if err != nil {
		return err
	}
	tc.log.Debug("copied remote state config into container")

	return nil
}

func (tc *TerraformContainer) Prepare() error {
	// build terraform image if needed
	tc.log.Debug("prepare container")

	// prepare environment
	environment := []string{}
	if environmentProvider, err := tc.t.context.ProviderEnvironment(); err != nil {
		return fmt.Errorf("error getting environment secrets from provider: %s", err)
	} else {
		environment = append(environment, environmentProvider...)
	}
	tc.log.WithField("environment", environment).Debug("")

	image, err := tc.t.DockerImage()
	if err != nil {
		return fmt.Errorf("error getting docker image failed: %s", err)
	}

	if tc.dockerContainer != nil {
		if err := tc.CleanUp(); err != nil {
			tc.log.WithField("container", tc.dockerContainer.ID).Warn("cleaning up of container failed")
		}
	}

	tc.dockerContainer, err = tc.t.dockerClient.CreateContainer(docker.CreateContainerOptions{
		Config: &docker.Config{
			Image: image.ID,
			Cmd: []string{
				"sleep",
				"3600",
			},
			WorkingDir: "/terraform",
			Env:        environment,
		},
	})
	if err != nil {
		return err
	}
	tc.log = tc.log.WithField("container", tc.dockerContainer.ID)

	tarOpts := &archive.TarOptions{
		Compression:  archive.Uncompressed,
		NoLchown:     true,
		IncludeFiles: []string{"."},
	}

	terraformDir := filepath.Clean(filepath.Join(tc.t.rootPath, "terraform/aws-centos", tc.stack.StackName()))
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

	err = tc.t.dockerClient.UploadToContainer(tc.dockerContainer.ID, docker.UploadToContainerOptions{
		InputStream: terraformTar,
		Path:        "/terraform",
	})
	if err != nil {
		return err
	}
	tc.log.Debug("copied terraform manifests into container")

	err = tc.t.dockerClient.StartContainer(tc.dockerContainer.ID, &docker.HostConfig{})
	if err != nil {
		return fmt.Errorf("error starting container '%s': %s", tc.dockerContainer.ID, err)
	}

	return nil
}
