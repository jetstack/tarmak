// Copyright Jetstack Ltd. See LICENSE for details.
package config

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/sirupsen/logrus"

	"github.com/jetstack/tarmak/pkg/tarmak/mocks"
)

type fakeConfig struct {
	*Config
	ctrl *gomock.Controller

	configPath string

	fakeTarmak *mocks.MockTarmak
}

func (f *fakeConfig) Finish() {
	if f.configPath != "" {
		os.RemoveAll(f.configPath)
		f.configPath = ""
	}
	f.ctrl.Finish()
}

func newFakeConfig(t *testing.T) *fakeConfig {
	c := &fakeConfig{
		ctrl: gomock.NewController(t),
	}

	// fakeTarmak
	c.fakeTarmak = mocks.NewMockTarmak(c.ctrl)

	// create temporary dir
	var err error
	c.configPath, err = ioutil.TempDir("", "tarmak-config")
	if err != nil {
		t.Fatal("cannot create temp dir: ", err)
	}
	c.fakeTarmak.EXPECT().ConfigPath().AnyTimes().Return(c.configPath)

	// setup custom logger
	logger := logrus.New()
	if testing.Verbose() {
		logger.Level = logrus.DebugLevel
	} else {
		logger.Out = ioutil.Discard
	}
	c.fakeTarmak.EXPECT().Log().AnyTimes().Return(logger.WithField("app", "tarmak"))

	c.Config, err = New(c.fakeTarmak)

	if err != nil {
		t.Error("unexpected error: ", err)
	}

	return c
}

func TestNewAmazonConfigClusterSingle(t *testing.T) {
	c := newFakeConfig(t)
	defer c.Finish()

	conf := c.NewAmazonConfigClusterSingle()

	err := c.writeYAML(conf)
	if err != nil {
		t.Error("unexpected error: ", err)
	}
	c.log.Infof("wrote configuration to: %s\n", c.Config.configPath())

	_, err = c.ReadConfig()
	if err != nil {
		t.Error("unexpected error: ", err)
	}

}
