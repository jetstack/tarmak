// Copyright Jetstack Ltd. See LICENSE for details.
package tags

import (
	"fmt"

	"github.com/jetstack/tarmak/pkg/wing/tags/aws"
)

type Tags interface {
	EnsureMachineTags() error
}

func New(targetProvider string) (Tags, error) {
	switch targetProvider {
	case "amazon", "aws":
		return aws.New(), nil

	default:
		return nil, fmt.Errorf("target provider for tags not supported %s", targetProvider)
	}
}
