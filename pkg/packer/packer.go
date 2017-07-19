package packer

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	logrus "github.com/Sirupsen/logrus"

	tarmakDocker "github.com/jetstack/tarmak/pkg/docker"
	"github.com/jetstack/tarmak/pkg/tarmak/interfaces"
)

const PackerTagEnvironment = "tarmak_environment"
const PackerTagBaseImageName = "tarmak_base_image_name"

type Packer struct {
	*tarmakDocker.App
	log    *logrus.Entry
	tarmak interfaces.Tarmak

	imageID *string
}

func New(tarmak interfaces.Tarmak) *Packer {
	log := tarmak.Log().WithField("module", "packer")

	app := tarmakDocker.NewApp(
		tarmak,
		log,
		"jetstack/tarmak-packer",
		"packer",
	)

	p := &Packer{
		App:    app,
		tarmak: tarmak,
		log:    log,
	}
	for key, val := range p.tags() {
		p.log = p.log.WithField(key, val)
	}

	return p
}

func (p *Packer) tags() map[string]string {
	return map[string]string{
		PackerTagEnvironment:   p.tarmak.Context().Environment().Name(),
		PackerTagBaseImageName: p.tarmak.Context().BaseImage(),
	}
}

func (p *Packer) QueryAMIID() (amiID string, err error) {
	if p.imageID != nil {
		return *p.imageID, nil
	}

	imageID, err := p.tarmak.Context().Environment().Provider().QueryImage(
		p.tags(),
	)

	p.imageID = &imageID

	return imageID, nil

}

func (p *Packer) Build() (amiID string, err error) {
	c := p.Container()

	rootPath, err := p.tarmak.RootPath()
	if err != nil {
		return "", fmt.Errorf("error getting rootPath: %s", err)
	}

	// set tarmak environment vars vars
	for key, value := range p.tags() {
		c.Env = append(c.Env, fmt.Sprintf("%s=%s", strings.ToUpper(key), value))
	}

	// get aws secrets
	if environmentProvider, err := p.tarmak.Context().Environment().Provider().Environment(); err != nil {
		return "", fmt.Errorf("error getting environment secrets from provider: %s", err)
	} else {
		c.Env = append(c.Env, environmentProvider...)
	}

	c.WorkingDir = "/packer"
	c.Cmd = []string{"sleep", "3600"}

	err = c.Prepare()
	if err != nil {
		return "", err
	}

	// make sure container get's cleaned up
	defer c.CleanUpSilent(p.log)

	buildSourcePath := filepath.Join(
		rootPath,
		"packer",
		fmt.Sprintf("%s.json", p.tarmak.Context().BaseImage),
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

	err = c.UploadToContainer(buildTar, "/packer")
	if err != nil {
		return "", err
	}
	p.log.Debug("copied packer build state")

	err = c.Start()
	if err != nil {
		return "", fmt.Errorf("error starting container: %s", err)
	}

	returnCode, err := c.Execute("packer", []string{"build", buildPath})
	if err != nil {
		return "", err
	}
	if exp, act := 0, returnCode; exp != act {
		return "", fmt.Errorf("unexpected return code: exp=%d, act=%d", exp, act)
	}

	return "unknown", nil
}
