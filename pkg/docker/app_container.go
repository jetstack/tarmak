package docker

import (
	"bufio"
	"fmt"
	"io"

	logrus "github.com/Sirupsen/logrus"
	"github.com/fsouza/go-dockerclient"
)

type AppContainer struct {
	log *logrus.Entry
	app *App

	dockerContainer *docker.Container

	WorkingDir string
	Env        []string
	Cmd        []string
}

func (ac *AppContainer) SetLog(log *logrus.Entry) {
	ac.log = log
}

func (ac *AppContainer) UploadToContainer(tarStream io.Reader, destPath string) error {
	return ac.app.dockerClient.UploadToContainer(ac.dockerContainer.ID, docker.UploadToContainerOptions{
		InputStream: tarStream,
		Path:        destPath,
	})
}

func (ac *AppContainer) Prepare() error {
	imageID, err := ac.app.ImageID()
	if err != nil {
		return fmt.Errorf("error getting docker image failed: %s", err)
	}

	if ac.dockerContainer != nil {
		if err := ac.CleanUp(); err != nil {
			ac.log.WithField("container", ac.dockerContainer.ID).Warn("cleaning up of container failed")
		}
	}

	ac.dockerContainer, err = ac.app.dockerClient.CreateContainer(docker.CreateContainerOptions{
		Config: &docker.Config{
			Image:      imageID,
			Cmd:        ac.Cmd,
			WorkingDir: ac.WorkingDir,
			Env:        ac.Env,
		},
	})
	if err != nil {
		return err
	}
	ac.log = ac.log.WithField("container", ac.dockerContainer.ID)

	return nil

}

func (ac *AppContainer) CleanUpSilent(log *logrus.Entry) {
	err := ac.CleanUp()
	if err != nil && log != nil {
		log.Warn("error cleaning up container: %s", err)
	}
}

func (ac *AppContainer) CleanUp() error {
	info, err := ac.app.dockerClient.InspectContainer(ac.dockerContainer.ID)
	if err != nil {
		return fmt.Errorf("error getting status of container %s: %s", ac.dockerContainer.ID, err)
	}

	if info.State.Running {
		err := ac.app.dockerClient.KillContainer(docker.KillContainerOptions{
			ID:     ac.dockerContainer.ID,
			Signal: docker.SIGKILL,
		})
		if err != nil {
			return fmt.Errorf("error sending KILL signal to container %s: %s", ac.dockerContainer.ID, err)
		}
	}

	return ac.app.dockerClient.RemoveContainer(docker.RemoveContainerOptions{ID: ac.dockerContainer.ID})
}

func (ac *AppContainer) Execute(cmd string, args []string) (returnCode int, err error) {
	command := []string{cmd}
	command = append(command, args...)
	ac.log.WithField("command", command).Debug()
	exec, err := ac.app.dockerClient.CreateExec(docker.CreateExecOptions{
		Cmd:          command,
		AttachStdin:  false,
		AttachStdout: true,
		AttachStderr: true,
		Container:    ac.dockerContainer.ID,
	})
	if err != nil {
		return -1, err
	}

	stdoutReader, stdoutWriter := io.Pipe()
	stdoutScanner := bufio.NewScanner(stdoutReader)
	go func() {
		for stdoutScanner.Scan() {
			ac.log.WithField("command", cmd).Debug(stdoutScanner.Text())
		}
	}()

	err = ac.app.dockerClient.StartExec(exec.ID, docker.StartExecOptions{
		ErrorStream:  stdoutWriter,
		OutputStream: stdoutWriter,
	})
	stdoutReader.Close()
	stdoutWriter.Close()

	if err != nil {
		return -1, fmt.Errorf("error starting exec: %s", err)
	}

	execInspect, err := ac.app.dockerClient.InspectExec(exec.ID)
	if err != nil {
		return -1, fmt.Errorf("error inspecting exec: %s", err)
	}

	return execInspect.ExitCode, nil
}

func (ac *AppContainer) Start() error {
	return ac.app.dockerClient.StartContainer(ac.dockerContainer.ID, &docker.HostConfig{})
}
