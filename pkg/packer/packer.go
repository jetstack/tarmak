package packer

import (
	logrus "github.com/Sirupsen/logrus"

	tarmakDocker "github.com/jetstack/tarmak/pkg/docker"
	"github.com/jetstack/tarmak/pkg/tarmak/config"
)

type Packer struct {
	app *tarmakDocker.App

	log    *logrus.Entry
	tarmak config.Tarmak
}

func New(t config.Tarmak) *Packer {
	p := &Packer{
		log:    t.Log().WithField("module", "packer"),
		tarmak: t,
	}

	return p
}
