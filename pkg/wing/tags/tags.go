// Copyright Jetstack Ltd. See LICENSE for details.
package tags

import (
	"errors"
	"fmt"
	"os"

	"github.com/jetstack/tarmak/pkg/wing/tags/aws"
	"github.com/sirupsen/logrus"
)

type Tags interface {
	EnsureMachineTags() error
}

func New(log *logrus.Entry) (Tags, error) {
	if log == nil {
		log = logrus.NewEntry(logrus.New())
		log.Level = logrus.DebugLevel
	}

	provider := os.Getenv("WING_CLOUD_PROVIDER")

	switch provider {
	case "amazon", "aws":
		return aws.New(log), nil

	case "":
		return nil, errors.New("no cloud provider specified, set environment variable WING_CLOUD_PROVIDER")

	default:
		return nil, fmt.Errorf("target provider for tags not supported %s", provider)
	}
}
