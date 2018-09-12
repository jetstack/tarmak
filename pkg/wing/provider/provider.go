// Copyright Jetstack Ltd. See LICENSE for details.
package provider

import (
	"io"

	"github.com/hashicorp/go-multierror"
	"github.com/sirupsen/logrus"

	"github.com/jetstack/tarmak/pkg/wing/provider/file"
	"github.com/jetstack/tarmak/pkg/wing/provider/hash"
	"github.com/jetstack/tarmak/pkg/wing/provider/s3"
)

type Provider interface {
	GetManifest(manifest string) (io.ReadCloser, error)
}

func GetManifest(log *logrus.Entry, manifestURL string) (io.ReadCloser, error) {
	var result *multierror.Error

	for _, p := range providers() {
		rc, err := p.GetManifest(manifestURL)
		if err != nil {
			result = multierror.Append(result, err)
			continue
		}

		return rc, nil
	}

	return nil, result.ErrorOrNil()
}

func providers() []Provider {
	return []Provider{
		&hash.Hash{},
		&s3.S3{},
		&file.File{},
	}
}
