package packer

import (
	"time"

	"github.com/Sirupsen/logrus"

	tarmakv1alpha1 "github.com/jetstack/tarmak/pkg/apis/tarmak/v1alpha1"
	tarmakDocker "github.com/jetstack/tarmak/pkg/docker"
	"github.com/jetstack/tarmak/pkg/tarmak/interfaces"
)

type Packer struct {
	*tarmakDocker.App
	log    *logrus.Entry
	tarmak interfaces.Tarmak
}

var _ interfaces.Packer = &Packer{}

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

// List necessary images for stack
func (p *Packer) images() (images []*image) {
	environment := p.tarmak.Cluster().Environment().Name()
	for _, imageName := range p.tarmak.Cluster().Images() {
		image := &image{
			environment: environment,
			imageName:   imageName,
			packer:      p,
			tarmak:      p.tarmak,
		}
		image.log = p.log
		for key, val := range image.tags() {
			image.log = image.log.WithField(key, val)
		}

		images = append(images, image)
	}

	return images
}

// List existing images
func (p *Packer) List() ([]tarmakv1alpha1.Image, error) {
	return p.tarmak.Cluster().Environment().Provider().QueryImages(
		map[string]string{tarmakv1alpha1.ImageTagEnvironment: p.tarmak.Environment().Name()},
	)
}

// Build all images
func (p *Packer) Build() error {
	for _, image := range p.images() {
		amiID, err := image.Build()
		if err != nil {
			return err
		}
		image.log.WithField("ami_id", amiID).Debugf("successfully built image")
	}
	return nil
}

// Query images
func (p *Packer) IDs() (map[string]string, error) {
	images, err := p.List()
	if err != nil {
		return nil, err
	}

	imagesChangeTime := make(map[string]time.Time)
	imageIDByName := make(map[string]string)

	for _, image := range images {
		if changeTime, ok := imagesChangeTime[image.BaseImage]; !ok || changeTime.Before(image.CreationTimestamp.Time) {
			imagesChangeTime[image.BaseImage] = image.CreationTimestamp.Time
			imageIDByName[image.BaseImage] = image.Name
		}
	}

	return imageIDByName, nil
}
