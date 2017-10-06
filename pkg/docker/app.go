// Copyright Jetstack Ltd. See LICENSE for details.
package docker

import (
	"bufio"
	"fmt"
	"io"
	"path/filepath"

	logrus "github.com/Sirupsen/logrus"
	"github.com/fsouza/go-dockerclient"

	"github.com/jetstack/tarmak/pkg/tarmak/interfaces"
)

type App struct {
	log    *logrus.Entry
	tarmak interfaces.Tarmak

	dockerClient *docker.Client
	dockerImage  *docker.Image

	imagePath string // relative image path to Dockerfile
	imageName string // docker image name

	DockerfilePath string
}

func NewApp(t interfaces.Tarmak, log *logrus.Entry, imageName, imagePath string) *App {
	return &App{
		imagePath: imagePath,
		imageName: imageName,
		log:       log,
		tarmak:    t,
	}
}

func (a *App) buildContextPath() (string, error) {
	rootPath, err := a.tarmak.RootPath()
	if err != nil {
		return "", err
	}

	return filepath.Join(rootPath, a.imagePath), nil
}

func (a *App) ImageID() (str string, err error) {
	if a.dockerImage != nil {
		return a.dockerImage.ID, nil
	}
	contextDir, err := a.buildContextPath()
	if err != nil {
		return "", fmt.Errorf("error building context path: %s", err)
	}

	a.log.Debugf("building image %s", a.imageName)
	a.dockerClient, err = docker.NewClientFromEnv()
	if err != nil {
		return "", err
	}

	stdoutReader, stdoutWriter := io.Pipe()
	stdoutScanner := bufio.NewScanner(stdoutReader)
	go func() {
		for stdoutScanner.Scan() {
			a.log.WithField("action", "docker-build").Debug(stdoutScanner.Text())
		}
	}()

	err = a.dockerClient.BuildImage(docker.BuildImageOptions{
		Name:         a.imageName,
		OutputStream: stdoutWriter,
		ContextDir:   contextDir,
	})
	stdoutReader.Close()
	stdoutWriter.Close()
	if err != nil {
		return "", err
	}

	image, err := a.dockerClient.InspectImage(a.imageName)
	if err != nil {
		return "", err
	}

	a.log.Debugf("successfully build image with id '%s'", image.ID)
	a.dockerImage = image
	return image.ID, nil
}

func (a *App) cleanupContainer(container *docker.Container) error {
	return nil
}

func (a *App) Cleanup() error {
	return nil
}

func (a *App) Container() *AppContainer {
	ac := &AppContainer{
		app: a,
		log: a.log.WithField("container", nil),
	}
	return ac
}
