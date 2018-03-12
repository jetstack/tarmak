// Copyright Jetstack Ltd. See LICENSE for details.
package docker

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"

	"github.com/fsouza/go-dockerclient"
	"github.com/jetstack/tarmak/pkg/tarmak/interfaces"
	"github.com/jetstack/tarmak/pkg/terraform/providers/tarmak/rpc"
	logrus "github.com/sirupsen/logrus"
)

type AppContainer struct {
	log *logrus.Entry
	app *App

	dockerContainer *docker.Container

	WorkingDir string   // working dir inside the container
	Env        []string // environment variables
	Cmd        []string // command to run in the container
	Keep       bool     // don't cleanup containers

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

	if ac.Keep {
		ac.log.Debug("keep container as requested")
		return nil
	}

	return ac.app.dockerClient.RemoveContainer(docker.RemoveContainerOptions{ID: ac.dockerContainer.ID})
}

func (ac *AppContainer) execute(cmd string, args []string, stdOut io.Writer, stdErr io.Writer, stdIn io.Reader, tty bool) (returnCode int, err error) {
	command := []string{cmd}
	command = append(command, args...)
	ac.log.WithField("command", command).Debug()
	exec, err := ac.app.dockerClient.CreateExec(docker.CreateExecOptions{
		Cmd:          command,
		AttachStdin:  stdIn != nil,
		AttachStdout: stdOut != nil,
		AttachStderr: stdErr != nil,
		Container:    ac.dockerContainer.ID,
		Tty:          tty,
	})
	if err != nil {
		return -1, err
	}

	err = ac.app.dockerClient.StartExec(exec.ID, docker.StartExecOptions{
		InputStream:  stdIn,
		ErrorStream:  stdErr,
		OutputStream: stdOut,
		Tty:          tty,
		RawTerminal:  tty,
	})

	if err != nil {
		return -1, fmt.Errorf("error starting exec: %s", err)
	}

	execInspect, err := ac.app.dockerClient.InspectExec(exec.ID)
	if err != nil {
		return -1, fmt.Errorf("error inspecting exec: %s", err)
	}

	return execInspect.ExitCode, nil
}

func (ac *AppContainer) Attach(cmd string, args []string) (returnCode int, err error) {
	return ac.execute(cmd, args, os.Stdout, os.Stderr, os.Stdin, true)
}

func (ac *AppContainer) Execute(cmd string, args []string) (returnCode int, err error) {
	stdoutReader, stdoutWriter := io.Pipe()
	stdoutScanner := bufio.NewScanner(stdoutReader)
	go func() {
		for stdoutScanner.Scan() {
			ac.log.WithField("command", cmd).Debug(stdoutScanner.Text())
		}
	}()
	returnCode, err = ac.execute(cmd, args, stdoutWriter, stdoutWriter, nil, false)
	stdoutReader.Close()
	stdoutWriter.Close()
	return returnCode, err
}

func (ac *AppContainer) Capture(cmd string, args []string) (stdOut string, stdErr string, returnCode int, err error) {
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
		return "", "", -1, err
	}

	var stdOutBuf bytes.Buffer
	var stdErrBuf bytes.Buffer
	stdOutWriter := bufio.NewWriter(&stdOutBuf)
	stdErrWriter := bufio.NewWriter(&stdErrBuf)

	err = ac.app.dockerClient.StartExec(exec.ID, docker.StartExecOptions{
		ErrorStream:  stdErrWriter,
		OutputStream: stdOutWriter,
	})
	if err != nil {
		return "", "", -1, fmt.Errorf("error starting exec: %s", err)
	}

	err = stdOutWriter.Flush()
	if err != nil {
		return "", "", -1, fmt.Errorf("error flushing stdout: %s", err)
	}
	err = stdErrWriter.Flush()
	if err != nil {
		return "", "", -1, fmt.Errorf("error flushing stdout: %s", err)
	}

	execInspect, err := ac.app.dockerClient.InspectExec(exec.ID)
	if err != nil {
		return "", "", -1, fmt.Errorf("error inspecting exec: %s", err)
	}

	return stdOutBuf.String(), stdErrBuf.String(), execInspect.ExitCode, nil
}

func (ac *AppContainer) Start() error {
	return ac.app.dockerClient.StartContainer(ac.dockerContainer.ID, &docker.HostConfig{})
}

// launch tarmak connector and attach an RPC server to it
func (ac *AppContainer) ListenRPC(tarmak interfaces.Tarmak, stack interfaces.Stack) (err error) {

	// create exec
	exec, err := ac.app.dockerClient.CreateExec(docker.CreateExecOptions{
		Cmd:          []string{"tarmak-connector"},
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,
		Container:    ac.dockerContainer.ID,
		Tty:          false,
	})
	if err != nil {
		return err
	}

	containerR, rpcW := io.Pipe()
	rpcR, containerW := io.Pipe()
	stderrReader, stderrWriter := io.Pipe()

	stderrScanner := bufio.NewScanner(stderrReader)
	go func() {
		for stderrScanner.Scan() {
			ac.log.WithField("app", "connector").Debug(stderrScanner.Text())
		}
	}()

	execWaiter, err := ac.app.dockerClient.StartExecNonBlocking(exec.ID, docker.StartExecOptions{
		InputStream:  containerR,
		ErrorStream:  stderrWriter,
		OutputStream: containerW,
	})
	if err != nil {
		return err
	}

	go func() {
		rpc.Bind(ac.log, rpc.NewTarmak(tarmak, stack), rpcR, rpcW, &execCloser{})
	}()

	go func() {
		if err := execWaiter.Wait(); err != nil {
			ac.log.Warnf("error waiting for termination of tarmak connector: %s", err)
		}

		state, err := ac.app.dockerClient.InspectExec(exec.ID)
		if err != nil {
			ac.log.Warnf("error getting status of tarmak connector: %s", err)
		}
		ac.log.Debugf("tarmak connector exited (code %d)", state.ExitCode)
	}()

	return nil
}

type execCloser struct {
}

func (c *execCloser) Close() error {
	// TODO: kill tarmak connector
	return nil
}
