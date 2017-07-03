package docker

import (
	"bufio"
	"io"
	"path/filepath"

	logrus "github.com/Sirupsen/logrus"
	"github.com/fsouza/go-dockerclient"

	"github.com/jetstack/tarmak/pkg/tarmak/config"
)

type App struct {
	log    *logrus.Entry
	tarmak config.Tarmak

	dockerClient *docker.Client
	dockerImage  *docker.Image

	imagePath string // relative image path to Dockerfile
	imageName string // docker image name

	DockerfilePath string
}

func NewApp(t config.Tarmak, log *logrus.Entry, imageName, imagePath string) *App {
	return &App{
		imagePath: imagePath,
		imageName: imageName,
		log:       log,
		tarmak:    t,
	}
}

func (a *App) buildContextPath() string {
	return filepath.Join(a.tarmak.RootPath(), a.imagePath)
}

func (a *App) ImageID() (str string, err error) {
	if a.dockerImage != nil {
		return a.dockerImage.ID, nil
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
		ContextDir:   a.buildContextPath(),
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
