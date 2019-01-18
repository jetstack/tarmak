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
	fileprovisioner "github.com/hashicorp/packer/provisioner/file"
	puppetmasterlessprovisioner "github.com/hashicorp/packer/provisioner/puppet-masterless"
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
	"shell":             new(shellprovisioner.Provisioner),
	"file":              new(fileprovisioner.Provisioner),
	"puppet-masterless": new(puppetmasterlessprovisioner.Provisioner),
}

type image struct {
	packer *Packer
	log    *logrus.Entry
	tarmak interfaces.Tarmak
	ctx    interfaces.CancellationContext

	environment string
	imageName   string
	id          *string
}

func (i *image) userVariables() map[string]string {
	return map[string]string{
		tarmakv1alpha1.ImageTagEnvironment:       i.environment,
		tarmakv1alpha1.ImageTagBaseImageName:     i.imageName,
		tarmakv1alpha1.ImageTagKubernetesVersion: i.tarmak.Cluster().Config().Kubernetes.Version,
		"region":                                 i.tarmak.Provider().Region(),
	}
}

func (i *image) Build() (amiID string, err error) {

	select {
	case <-i.ctx.Done():
		return "", i.ctx.Err()
	default:
	}

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
		Variables:  i.userVariables(),
	}

	select {
	case <-i.ctx.Done():
		return "", i.ctx.Err()
	default:
	}

	envVars, err := i.tarmak.Provider().Environment()
	if err != nil {
		return "", fmt.Errorf("failed to get provider credentials: %v", err)
	}
	rootPath, err = i.tarmak.RootPath()
	if err != nil {
		return "", fmt.Errorf("error getting rootPath: %s", err)
	}
	path := filepath.Join(rootPath, "puppet")

	envVars = append(envVars, fmt.Sprintf("PUPPET_PATH=%s", path))

	var result *multierror.Error
	for _, e := range envVars {
		kv := strings.Split(e, "=")
		if len(kv) < 2 {
			err := fmt.Errorf("malformed environment variable: %s", kv)
			result = multierror.Append(result, err)
			continue
		}

		v := strings.Join(kv[1:], "")
		if err := os.Setenv(kv[0], v); err != nil {
			result = multierror.Append(result, err)
		}
	}

	if result.ErrorOrNil() != nil {
		return "", result.ErrorOrNil()
	}

	select {
	case <-i.ctx.Done():
		return "", i.ctx.Err()
	default:
	}

	core, err := packer.NewCore(config)
	if err != nil {
		return "", fmt.Errorf("failed to get core: %v", err)
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	var amiIDs []string
	var builds []packer.Build

	for _, buildName := range core.BuildNames() {
		build, err := core.Build(buildName)
		if err != nil {
			result = multierror.Append(result, err)
			continue
		}
		build.SetDebug(true)
		builds = append(builds, build)

		select {
		case <-i.ctx.Done():
			break
		default:
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
			complete := make(chan struct{})

			go func() {
				select {
				case <-i.ctx.Done():
					i.log.Warnf("attempting to cancel build '%s', please be patient while the builder shuts down.", build.Name())
					build.Cancel()
					<-complete

				case <-complete:
				}
			}()

			artifacts, err := build.Run(ui, nil)
			if err != nil {
				mu.Lock()
				result = multierror.Append(result, err)
				mu.Unlock()

				close(complete)
				return
			}

			for _, a := range artifacts {
				mu.Lock()
				amiIDs = append(amiIDs, a.Id())
				mu.Unlock()
			}

			close(complete)
			return
		}()
	}

	wg.Wait()

	return strings.Join(amiIDs, ", "), result.ErrorOrNil()
}
