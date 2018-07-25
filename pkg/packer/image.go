// Copyright Jetstack Ltd. See LICENSE for details.
package packer

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/hashicorp/go-multierror"
	amazonebsbuilder "github.com/hashicorp/packer/builder/amazon/ebs"
	"github.com/hashicorp/packer/packer"
	shellprovisioner "github.com/hashicorp/packer/provisioner/shell"
	"github.com/hashicorp/packer/template"
	"github.com/hashicorp/packer/version"
	logrus "github.com/sirupsen/logrus"

	tarmakv1alpha1 "github.com/jetstack/tarmak/pkg/apis/tarmak/v1alpha1"
	"github.com/jetstack/tarmak/pkg/tarmak/interfaces"
)

var Builders = map[string]packer.Builder{
	"amazon-ebs": new(amazonebsbuilder.Builder),
}

var Provisioners = map[string]packer.Provisioner{
	"shell": new(shellprovisioner.Provisioner),
}

type image struct {
	packer *Packer
	log    *logrus.Entry
	tarmak interfaces.Tarmak

	environment string
	imageName   string
	id          *string
}

func (i *image) tags() map[string]string {
	return map[string]string{
		tarmakv1alpha1.ImageTagEnvironment:   i.environment,
		tarmakv1alpha1.ImageTagBaseImageName: i.imageName,
		"region": i.tarmak.Provider().Region(),
	}
}

func (i *image) Build() (amiID string, err error) {
	rootPath, err := i.tarmak.RootPath()
	if err != nil {
		return "", fmt.Errorf("error getting rootPath: %s", err)
	}

	buildSourcePath := filepath.Join(
		rootPath,
		"packer",
		i.tarmak.Cluster().Environment().Provider().Cloud(),
		fmt.Sprintf("%s.json", i.imageName),
	)

	tpl, err := template.ParseFile(buildSourcePath)
	if err != nil {
		return "", fmt.Errorf("failed to parse template source file: %v", err)
	}

	components := packer.ComponentFinder{
		Builder: func(n string) (packer.Builder, error) {
			b, ok := Builders[n]
			if !ok {
				return nil, fmt.Errorf("builder '%s' not supported", n)
			}

			return b, nil
		},
		Provisioner: func(n string) (packer.Provisioner, error) {
			p, ok := Provisioners[n]
			if !ok {
				return nil, fmt.Errorf("provisioner '%s' not supported", n)
			}

			return p, nil
		},
	}

	config := &packer.CoreConfig{
		Version:    version.Version,
		Template:   tpl,
		Components: components,
		Variables:  i.tags(),
	}

	creds, err := i.tarmak.Provider().Credentials()
	if err != nil {
		return "", fmt.Errorf("faild to get provider credentials: %v", err)
	}

	var result *multierror.Error
	for k, v := range creds {
		if err := os.Setenv(k, v); err != nil {
			result = multierror.Append(result, err)
		}
	}

	if result.ErrorOrNil() != nil {
		return "", result.ErrorOrNil()
	}

	core, err := packer.NewCore(config)
	if err != nil {
		return "", fmt.Errorf("failed to get core: %v", err)
	}

	var wg sync.WaitGroup
	var amiIDs []string
	for _, buildName := range core.BuildNames() {
		build, err := core.Build(buildName)
		if err != nil {
			result = multierror.Append(result, err)
			continue
		}

		if _, err := build.Prepare(); err != nil {
			result = multierror.Append(result, err)
			continue
		}

		ui := &packer.ColoredUi{
			Ui: &packer.MachineReadableUi{
				Writer: i.log.Writer(),
			},
			ErrorColor: packer.UiColorRed,
			Color:      packer.UiColorBlue,
		}

		wg.Add(1)
		go func() {
			defer wg.Done()

			artifacts, err := build.Run(ui, nil)
			if err != nil {
				result = multierror.Append(result, err)
				return
			}

			for _, a := range artifacts {
				amiIDs = append(amiIDs, a.Id())
			}
		}()
	}

	wg.Wait()

	return strings.Join(amiIDs, ", "), result.ErrorOrNil()
}
