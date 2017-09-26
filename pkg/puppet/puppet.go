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

type hieraData struct {
	classes   []string
	variables []string
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

func kubernetesClusterConfig(conf *clusterv1alpha1.ClusterKubernetes, hieraData *hieraData) {
	if conf == nil {
		return
	}
	if conf.Version != "" {
		hieraData.variables = append(hieraData.variables, fmt.Sprintf(`tarmak::kubernetes_version: "%s"`, conf.Version))
	}
	return
}

func kubernetesClusterConfigPerRole(conf *clusterv1alpha1.ClusterKubernetes, roleName string, hieraData *hieraData) {
	if conf == nil {
		return
	}

	if roleName == clusterv1alpha1.KubernetesMasterRoleName && conf.ClusterAutoscaler != nil && conf.ClusterAutoscaler.Enabled {
		hieraData.classes = append(hieraData.classes, `kubernetes_addons::cluster_autoscaler`)
		hieraData.variables = append(hieraData.variables, fmt.Sprintf(`kubernetes_addons::cluster_autoscaler_image: "%s"`, conf.ClusterAutoscaler.Image))
		hieraData.variables = append(hieraData.variables, fmt.Sprintf(`kubernetes_addons::cluster_autoscaler_version: "%s"`, conf.ClusterAutoscaler.Version))
	}
	return
}

func kubernetesInstancePoolConfig(conf *clusterv1alpha1.InstancePoolKubernetes, hieraData *hieraData) {
	if conf == nil {
		return
	}
	if conf.Version != "" {
		hieraData.variables = append(hieraData.variables, fmt.Sprintf(`tarmak::kubernetes_version: "%s"`, conf.Version))
	}
	return
}

func contentClusterConfig(conf *clusterv1alpha1.Cluster) (lines []string) {

	hieraData := &hieraData{}
	kubernetesClusterConfig(conf.Kubernetes, hieraData)

	return serialiseHieraData(hieraData)
}

func contentInstancePoolConfig(clusterConf *clusterv1alpha1.Cluster, instanceConf *clusterv1alpha1.InstancePool, roleName string) (lines []string) {

	hieraData := &hieraData{}
	kubernetesClusterConfigPerRole(clusterConf.Kubernetes, roleName, hieraData)
	kubernetesInstancePoolConfig(instanceConf.Kubernetes, hieraData)

	return serialiseHieraData(hieraData)
}

func serialiseHieraData(hieraData *hieraData) (lines []string) {

	if hieraData == nil {
		return lines
	}

	if len(hieraData.classes) > 0 {
		lines = append(lines, `---`)
		lines = append(lines, `classes:`)
		for _, class := range hieraData.classes {
			lines = append(lines, fmt.Sprintf(`- %s`, class))
		}
	}

	for _, variable := range hieraData.variables {
		lines = append(lines, fmt.Sprintf(`- %s`, variable))
	}

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
