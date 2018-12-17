// Copyright Jetstack Ltd. See LICENSE for details.
package packer

import (
	"sync"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/sirupsen/logrus"

	tarmakv1alpha1 "github.com/jetstack/tarmak/pkg/apis/tarmak/v1alpha1"
	"github.com/jetstack/tarmak/pkg/tarmak/interfaces"
)

type Packer struct {
	log    *logrus.Entry
	tarmak interfaces.Tarmak
	ctx    interfaces.CancellationContext
}

var _ interfaces.Packer = &Packer{}

func New(tarmak interfaces.Tarmak) *Packer {
	log := tarmak.Log().WithField("module", "packer")

	p := &Packer{
		tarmak: tarmak,
		log:    log,
		ctx:    tarmak.CancellationContext(),
	}

	return p
}

// List existing images
func (p *Packer) List() ([]tarmakv1alpha1.Image, error) {
	return p.tarmak.Cluster().Environment().Provider().QueryImages(
		map[string]string{tarmakv1alpha1.ImageTagEnvironment: p.tarmak.Environment().Name()},
	)
}

// Build images
func (p *Packer) Build(imageNames []string) error {
	p.log.Infof("building images %s", imageNames)

	var resultLock sync.Mutex
	var wg sync.WaitGroup
	var result *multierror.Error

	// Save puppet config
	err := p.tarmak.Puppet().Initialize(true)
	if err != nil {
		return err
	}
	wg.Add(len(imageNames))
	for _, name := range imageNames {
		image := &image{
			environment: p.tarmak.Environment().Name(),
			imageName:   name,
			packer:      p,
			tarmak:      p.tarmak,
			ctx:         p.tarmak.CancellationContext(),
		}
		image.log = p.log
		for key, val := range image.userVariables() {
			image.log = image.log.WithField(key, val)
		}

		go func() {
			resultLock.Lock()
			amiID, err := image.Build()

			defer wg.Done()
			defer resultLock.Unlock()

			if err != nil {
				result = multierror.Append(result, err)
				return
			}

			// ensure we catch the interrupt error
			select {
			case <-p.ctx.Done():
				result = multierror.Append(result, p.ctx.Err())
				return
			default:
			}

			image.log.WithField("ami_id", amiID).Infof("successfully built image %s", image.imageName)
		}()

	}

	wg.Wait()

	return result.ErrorOrNil()
}

// Query images
func (p *Packer) IDs(encrypted bool) (map[string]string, error) {
	images, err := p.List()
	if err != nil {
		return nil, err
	}

	imagesChangeTime := make(map[string]time.Time)
	imageIDByName := make(map[string]string)

	for _, image := range images {
		if image.Encrypted != encrypted {
			continue
		}
		if changeTime, ok := imagesChangeTime[image.BaseImage]; !ok || changeTime.Before(image.CreationTimestamp.Time) {
			imagesChangeTime[image.BaseImage] = image.CreationTimestamp.Time
			imageIDByName[image.BaseImage] = image.Name
		}
	}

	return imageIDByName, nil
}
