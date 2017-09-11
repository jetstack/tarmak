package packer

import (
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

	return p
}

func (p *Packer) images() (images []*image) {
	environment := p.tarmak.Context().Environment().Name()
	for _, imageName := range p.tarmak.Context().Images() {
		image := &image{
			environment: environment,
			imageName:   imageName,
		}
		image.log = p.log
		for key, val := range image.tags() {
			image.log = image.log.WithField(key, val)
		}

		images = append(images, image)
	}

	return images
}
