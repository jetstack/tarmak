package terraform

import (
	"bufio"
	"fmt"
	"io"

	log "github.com/Sirupsen/logrus"
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
	dockerClient *docker.Client
	dockerImage  *docker.Image

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

func (t *Terraform) cleanupContainer(container *docker.Container) error {
	return nil
}

func (t *Terraform) Cleanup() error {
	return nil
}

func (t *Terraform) NewContainer(stack *config.Stack) *TerraformContainer {
	c := &TerraformContainer{
		t:     t,
		log:   t.log.WithField("stack", stack.StackName()),
		stack: stack,
	}
	return c
}

func (t *Terraform) Apply(stack *config.Stack) error {
	return t.planApply(stack, false)
}

func (t *Terraform) Destroy(stack *config.Stack) error {
	return t.planApply(stack, true)
}

func (t *Terraform) planApply(stack *config.Stack, destroy bool) error {
	c := t.NewContainer(stack)

	if err := c.Prepare(); err != nil {
		return fmt.Errorf("error preparing container: %s", err)
	}

	initialStateStack := false
	// check for initial state run on first deployment
	if !destroy && stack.StackName() == config.StackNameState {
		remoteStateAvail, err := t.context.RemoteStateAvailable()
		if err != nil {
			return fmt.Errorf("error finding remote state: %s", err)
		}
		if !remoteStateAvail {
			initialStateStack = true
			c.log.Infof("running state stack for the first time, by passing remote state")
		}
	}

	if !initialStateStack {
		err := c.CopyRemoteState(t.context.RemoteState(stack.StackName()))

		if err != nil {
			return fmt.Errorf("error while copying remote state: %s", err)
		}
		c.log.Debug("copied remote state into container")
	}

	if err := c.Init(); err != nil {
		return fmt.Errorf("error while terraform init: %s", err)
	}

	// check for destroying the state stack
	if destroy && stack.StackName() == config.StackNameState {
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

	changesNeeded, err := c.Plan(destroy)
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
		err := c.CopyRemoteState(t.context.RemoteState(stack.StackName()))
		if err != nil {
			return fmt.Errorf("error while copying remote state: %s", err)
		}
		c.log.Debug("copied remote state into container")

		if err := c.InitForceCopy(); err != nil {
			return fmt.Errorf("error while terraform init -force-copy: %s", err)
		}
	}

	return nil
}
