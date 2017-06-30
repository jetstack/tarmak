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

var terraformDockerImageName = "jetstack/tarmak-terraform"

var terraformDockerfile = `
FROM alpine:3.6

RUN apk add --no-cache unzip curl

# install terraform
ENV TERRAFORM_VERSION 0.9.8
ENV TERRAFORM_HASH f951885f4e15deb4cf66f3b199964e3e74a0298bb46c9fe42e105df2ebcf3d16
RUN curl -sL  https://releases.hashicorp.com/terraform/${TERRAFORM_VERSION}/terraform_${TERRAFORM_VERSION}_linux_amd64.zip > /tmp/terraform.zip && \
    echo "${TERRAFORM_HASH}  /tmp/terraform.zip" | sha256sum  -c && \
    unzip /tmp/terraform.zip && \
    rm /tmp/terraform.zip && \
    mv terraform /usr/local/bin/terraform && \
    chmod +x /usr/local/bin/terraform
`

type Terraform struct {
	dockerClient    *docker.Client
	dockerImage     *docker.Image
	dockerContainer *docker.Container

	rootPath string

	log *log.Entry

	context *config.Context
}

func New(logger *log.Entry, context *config.Context) *Terraform {
	if logger == nil {
		myLog := log.New()
		myLog.Level = log.DebugLevel
		logger = myLog.WithField("module", "terraform")
	}

	logger.Debug("initialising")

	return &Terraform{
		rootPath: "/home/christian/.golang/packages/src/github.com/jetstack/tarmak",
		log:      logger,
		context:  context,
	}
}

func (t *Terraform) DockerImage() (image *docker.Image, err error) {
	if t.dockerImage != nil {
		return t.dockerImage, nil
	}

	t.log.Debug("building terraform image")
	t.dockerClient, err = docker.NewClientFromEnv()
	if err != nil {
		return nil, err
	}

	dockerFile, err := tarmakDocker.TarStreamFromDockerfile(terraformDockerfile)
	if err != nil {
		return nil, err
	}

	stdoutReader, stdoutWriter := io.Pipe()
	stdoutScanner := bufio.NewScanner(stdoutReader)
	go func() {
		for stdoutScanner.Scan() {
			t.log.WithField("action", "docker-build").Debug(stdoutScanner.Text())
		}
	}()

	err = t.dockerClient.BuildImage(docker.BuildImageOptions{
		Name:         terraformDockerImageName,
		OutputStream: stdoutWriter,
		InputStream:  dockerFile,
	})
	stdoutReader.Close()
	stdoutWriter.Close()
	if err != nil {
		return nil, err
	}

	image, err = t.dockerClient.InspectImage(terraformDockerImageName)
	if err != nil {
		return nil, err
	}

	return image, nil

}

func (t *Terraform) prepareContainer(stack *config.Stack) error {
	remoteState := ""

	stackName := stack.StackName()

	logger := t.log.WithField("stack", stackName)
	logger.Debug("prepare new terraform container")

	image, err := t.DockerImage()
	if err != nil {
		return fmt.Errorf("error building of docker image failed: %s", err)
	}

	if t.dockerContainer != nil {
		t.cleanupContainer(t.dockerContainer)
		logger.WithField("container", t.dockerContainer.ID).Warn("cleaning up of container failed")
	}

	environment := []string{}

	if environmentProvider, err := t.context.ProviderEnvironment(); err != nil {
		return fmt.Errorf("error getting environment secrets from provider: %s", err)
	} else {
		environment = append(environment, environmentProvider...)
	}

	logger.WithField("environment", environment).Debug("")
	t.dockerContainer, err = t.dockerClient.CreateContainer(docker.CreateContainerOptions{
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
	logger = logger.WithField("container", t.dockerContainer.ID)

	tarOpts := &archive.TarOptions{
		Compression:  archive.Uncompressed,
		NoLchown:     true,
		IncludeFiles: []string{"."},
	}

	terraformDir := filepath.Clean(filepath.Join(t.rootPath, "terraform/aws-centos", stackName))
	logger = logger.WithField("terraform-dir", terraformDir)

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

	err = t.dockerClient.UploadToContainer(t.dockerContainer.ID, docker.UploadToContainerOptions{
		InputStream: terraformTar,
		Path:        "/terraform",
	})
	if err != nil {
		return err
	}
	logger.Debug("copied terraform manifests into container")

	remoteStateTar, err := tarmakDocker.TarStreamFromFile("terraform_remote_state.tf", remoteState)
	if err != nil {
		return err
	}

	err = t.dockerClient.UploadToContainer(t.dockerContainer.ID, docker.UploadToContainerOptions{
		InputStream: remoteStateTar,
		Path:        "/terraform",
	})
	if err != nil {
		return err
	}
	logger.Debug("copied remote state config into container")

	logger.Info("initialising terraform")

	err = t.dockerClient.StartContainer(t.dockerContainer.ID, &docker.HostConfig{})
	if err != nil {
		return fmt.Errorf("error starting container '%s': %s", t.dockerContainer.ID, err)
	}

	exec, err := t.dockerClient.CreateExec(docker.CreateExecOptions{
		Cmd:          []string{"terraform", "init"},
		AttachStdin:  false,
		AttachStdout: true,
		AttachStderr: true,
		Container:    t.dockerContainer.ID,
	})
	if err != nil {
		return err
	}

	stdoutReader, stdoutWriter := io.Pipe()
	stdoutScanner := bufio.NewScanner(stdoutReader)
	go func() {
		for stdoutScanner.Scan() {
			t.log.WithField("action", "terraform-init").Debug(stdoutScanner.Text())
		}
	}()

	err = t.dockerClient.StartExec(exec.ID, docker.StartExecOptions{
		ErrorStream:  stdoutWriter,
		OutputStream: stdoutWriter,
	})
	stdoutReader.Close()
	stdoutWriter.Close()
	if err != nil {
		return fmt.Errorf("error starting exec: %s", err)
	}

	execInspect, err := t.dockerClient.InspectExec(exec.ID)
	if err != nil {
		return fmt.Errorf("error inspecting exec: %s", err)
	}

	if execInspect.ExitCode != 0 {
		return fmt.Errorf("error initializing terraform failed: exit_code=%d", execInspect.ExitCode)
	}

	return nil
}

func (t *Terraform) cleanupContainer(container *docker.Container) error {
	return nil
}

func (t *Terraform) Cleanup() error {
	return nil
}

func (t *Terraform) Plan(stack *config.Stack) error {

	return t.prepareContainer(stack)

	return nil
}
