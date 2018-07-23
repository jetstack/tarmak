// Copyright Jetstack Ltd. See LICENSE for details.
package packer

import (
	"context"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	logrus "github.com/sirupsen/logrus"

	tarmakv1alpha1 "github.com/jetstack/tarmak/pkg/apis/tarmak/v1alpha1"
	tarmakDocker "github.com/jetstack/tarmak/pkg/docker"
	"github.com/jetstack/tarmak/pkg/tarmak/interfaces"
)

type image struct {
	packer *Packer
	log    *logrus.Entry
	tarmak interfaces.Tarmak

	environment string
	imageName   string
	id          *string
	ctx         context.Context
}

func (i *image) tags() map[string]string {
	return map[string]string{
		tarmakv1alpha1.ImageTagEnvironment:   i.environment,
		tarmakv1alpha1.ImageTagBaseImageName: i.imageName,
	}
}

func (i *image) Build(ctx context.Context) (amiID string, err error) {
	i.ctx = ctx

	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	c := i.packer.Container()

	rootPath, err := i.tarmak.RootPath()
	if err != nil {
		return "", fmt.Errorf("error getting rootPath: %s", err)
	}

	// set tarmak environment vars vars
	for key, value := range i.tags() {
		c.Env = append(c.Env, fmt.Sprintf("%s=%s", strings.ToUpper(key), value))
	}

	// get aws secrets
	if environmentProvider, err := i.tarmak.Cluster().Environment().Provider().Environment(); err != nil {
		return "", fmt.Errorf("error getting environment secrets from provider: %s", err)
	} else {
		c.Env = append(c.Env, environmentProvider...)
	}

	c.WorkingDir = "/packer"
	c.Cmd = []string{"sleep", "3600"}
	c.Keep = i.packer.tarmak.KeepContainers()

	complete := make(chan struct{})
	go i.waitOrKill(c, complete)

	err = c.Prepare()
	close(complete)
	if err != nil {
		return "", err
	}

	// make sure container get's cleaned up
	defer c.CleanUpSilent(i.log)

	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	buildSourcePath := filepath.Join(
		rootPath,
		"packer",
		i.tarmak.Cluster().Environment().Provider().Cloud(),
		fmt.Sprintf("%s.json", i.imageName),
	)

	buildContent, err := ioutil.ReadFile(buildSourcePath)
	if err != nil {
		return "", err
	}

	buildPath := "build.json"

	buildTar, err := tarmakDocker.TarStreamFromFile(buildPath, string(buildContent))
	if err != nil {
		return "", err
	}

	complete = make(chan struct{})
	go i.waitOrKill(c, complete)

	err = c.UploadToContainer(buildTar, "/packer")
	close(complete)
	if err != nil {
		return "", err
	}
	i.log.Debug("copied packer build state")

	complete = make(chan struct{})
	go i.waitOrKill(c, complete)

	err = c.Start()
	close(complete)
	if err != nil {
		return "", fmt.Errorf("error starting container: %s", err)
	}

	complete = make(chan struct{})
	go i.waitOrKill(c, complete)

	returnCode, err := c.Execute("packer", []string{"build", buildPath})
	if err != nil {
		return "", err
	}
	if exp, act := 0, returnCode; exp != act {
		return "", fmt.Errorf("packer returned unexpected return code: exp=%d, act=%d", exp, act)
	}
	close(complete)

	return "unknown", nil
}

func (i *image) waitOrKill(c *tarmakDocker.AppContainer, complete chan struct{}) {
	select {
	case <-i.ctx.Done():
		c.Kill()
		return
	case <-complete:
		return
	}
}
