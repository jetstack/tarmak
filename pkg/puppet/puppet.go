package puppet

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/docker/docker/pkg/archive"

	clusterv1alpha1 "github.com/jetstack/tarmak/pkg/apis/cluster/v1alpha1"
	tarmakv1alpha1 "github.com/jetstack/tarmak/pkg/apis/tarmak/v1alpha1"
	"github.com/jetstack/tarmak/pkg/tarmak/interfaces"
	"github.com/jetstack/tarmak/pkg/tarmak/utils"
)

type Puppet struct {
	log    *logrus.Entry
	tarmak interfaces.Tarmak
}

func New(tarmak interfaces.Tarmak) *Puppet {
	log := tarmak.Log().WithField("module", "puppet")

	return &Puppet{
		log:    log,
		tarmak: tarmak,
	}
}

func (p *Puppet) TarGz(writer io.Writer) error {

	rootPath, err := p.tarmak.RootPath()
	if err != nil {
		return fmt.Errorf("error getting rootPath: %s", err)
	}

	path := filepath.Join(rootPath, "puppet")

	err = p.writeHieraData(path, p.tarmak.Context())
	if err != nil {
		return err
	}

	// use same creation/mod time for all directories and files
	err = filepath.Walk(path, func(path string, f os.FileInfo, err error) error {
		return os.Chtimes(path, tarmakv1alpha1.KubernetesEpoch, tarmakv1alpha1.KubernetesEpoch)
	})
	if err != nil {
		return err
	}

	reader, err := archive.Tar(
		path,
		archive.Gzip,
	)
	if err != nil {
		return fmt.Errorf("error creating tar from path '%s': %s", path, err)
	}

	if _, err := io.Copy(writer, reader); err != nil {
		return fmt.Errorf("error writing tar: %s", err)
	}

	return nil
}

func kubernetesConfig(conf *clusterv1alpha1.Kubernetes) (lines []string) {
	if conf == nil {
		return lines
	}
	if conf.Version != "" {
		lines = append(lines, fmt.Sprintf(`tarmak::kubernetes_version: "%s"`, conf.Version))
	}

	return lines
}

func contentGlobalConfig(conf *clusterv1alpha1.Cluster) (lines []string) {
	lines = append(lines, kubernetesConfig(conf.Kubernetes)...)
	return lines
}

func contentInstancePoolConfig(conf *clusterv1alpha1.ServerPool) (lines []string) {
	lines = append(lines, kubernetesConfig(conf.Kubernetes)...)
	return lines
}

func (p *Puppet) writeLines(filePath string, lines []string) error {
	if len(lines) == 0 {
		// TODO: delete a potentially existing file
		return nil
	}
	err := utils.EnsureDirectory(filepath.Dir(filePath), 0750)
	if err != nil {
		return err
	}
	err = os.Chtimes(filepath.Dir(filePath), tarmakv1alpha1.KubernetesEpoch, tarmakv1alpha1.KubernetesEpoch)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(filePath, []byte(strings.Join(lines, "\n")), 0640)
	if err != nil {
		return err
	}
	return os.Chtimes(filePath, tarmakv1alpha1.KubernetesEpoch, tarmakv1alpha1.KubernetesEpoch)
}

func (p *Puppet) writeHieraData(puppetPath string, context interfaces.Context) error {

	hieraPath := filepath.Join(
		puppetPath,
		"hieradata",
	)

	// write global cluster config
	err := p.writeLines(
		filepath.Join(hieraPath, "tarmak.yaml"),
		contentGlobalConfig(context.Config()),
	)
	if err != nil {
		return fmt.Errorf("error writing global hiera config: %s", err)
	}

	// loop through instance pools
	for _, instancePool := range context.NodeGroups() {
		err = p.writeLines(
			filepath.Join(hieraPath, "instance_pools", fmt.Sprintf("%s.yaml", instancePool.Name())),
			contentInstancePoolConfig(instancePool.Config()),
		)
		if err != nil {
			return fmt.Errorf("error writing global hiera for instancePool %s: %s", instancePool.Name(), err)
		}
	}

	return nil

}
