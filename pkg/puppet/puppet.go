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

const (
	Master = "master"
	WORKER = "worker"
	ETC    = "etc"
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

	err = p.writeHieraData(path, p.tarmak.Cluster())
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

func kubernetesClusterConfig(conf *clusterv1alpha1.ClusterKubernetes) (lines []string) {
	if conf == nil {
		return lines
	}
	if conf.Version != "" {
		lines = append(lines, fmt.Sprintf(`tarmak::kubernetes_version: "%s"`, conf.Version))
	}

	return lines
}

func kubernetesClusterConfigPerRole(conf *clusterv1alpha1.ClusterKubernetes, role string) (lines []string) {
	if conf == nil {
		return lines
	}

	if role == clusterv1alpha1.KubernetesMasterRole && conf.ClusterAutoscaler != nil && conf.ClusterAutoscaler.Enabled {
		lines = append(lines, `classes:`)
		lines = append(lines, `- kubernetes_addons::cluster_autoscaler`)
		lines = append(lines, fmt.Sprintf(`tarmak::cluster_autoscaler_image: "%s"`, conf.ClusterAutoscaler.Image))
		lines = append(lines, fmt.Sprintf(`tarmak::cluster_autoscaler_version: "%s"`, conf.ClusterAutoscaler.Version))
	}

	return lines
}

func kubernetesInstancePoolConfig(conf *clusterv1alpha1.InstancePoolKubernetes) (lines []string) {
	if conf == nil {
		return lines
	}
	if conf.Version != "" {
		lines = append(lines, fmt.Sprintf(`tarmak::kubernetes_version: "%s"`, conf.Version))
	}

	return lines
}

func contentClusterConfig(conf *clusterv1alpha1.Cluster) (lines []string) {
	lines = append(lines, kubernetesClusterConfig(conf.Kubernetes)...)
	return lines
}

func contentInstancePoolConfig(clusterConf *clusterv1alpha1.Cluster, instanceConf *clusterv1alpha1.InstancePool, role string) (lines []string) {
	lines = append(lines, kubernetesClusterConfigPerRole(clusterConf.Kubernetes, role)...)
	lines = append(lines, kubernetesInstancePoolConfig(instanceConf.Kubernetes)...)
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

func (p *Puppet) writeHieraData(puppetPath string, cluster interfaces.Cluster) error {

	hieraPath := filepath.Join(
		puppetPath,
		"hieradata",
	)

	// write cluster config
	err := p.writeLines(
		filepath.Join(hieraPath, "tarmak.yaml"),
		contentClusterConfig(cluster.Config()),
	)
	if err != nil {
		return fmt.Errorf("error writing global hiera config: %s", err)
	}

	// loop through instance pools
	for _, instancePool := range cluster.InstancePools() {
		err = p.writeLines(
			filepath.Join(hieraPath, "instance_pools", fmt.Sprintf("%s.yaml", instancePool.Name())),
			contentInstancePoolConfig(cluster.Config(), instancePool.Config(), instancePool.Role().Name()),
		)
		if err != nil {
			return fmt.Errorf("error writing global hiera for instancePool %s: %s", instancePool.Name(), err)
		}
	}

	return nil

}
