package tarmak

import (
	"io/ioutil"
	"testing"

	"github.com/Sirupsen/logrus"

	"github.com/jetstack/tarmak/pkg/tarmak/config"
)

func newTarmak() *Tarmak {
	logger := logrus.New()
	if testing.Verbose() {
		logger.Level = logrus.DebugLevel
	} else {
		logger.Out = ioutil.Discard
	}
	return &Tarmak{
		log: logger,
	}
}

func TestTarmak_initFromConfig_DefaultHub(t *testing.T) {
	tarmak := newTarmak()
	err := tarmak.initFromConfig(config.DefaultConfigHub())
	if err != nil {
		t.Error("unexpected error: ", err)
	}
}

func TestTarmak_initFromConfig_DefaultSingle(t *testing.T) {
	tarmak := newTarmak()
	err := tarmak.initFromConfig(config.DefaultConfigSingle())
	if err != nil {
		t.Error("unexpected error: ", err)
	}
}

func TestTarmak_initFromConfig_DefaultSingleFrankfurt(t *testing.T) {
	tarmak := newTarmak()
	err := tarmak.initFromConfig(config.DefaultConfigSingleEnvSingleZoneAWSEUCentral())
	if err != nil {
		t.Error("unexpected error: ", err)
	}
}
