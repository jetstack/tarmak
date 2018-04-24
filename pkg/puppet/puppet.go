// Copyright Jetstack Ltd. See LICENSE for details.
package puppet

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/docker/docker/pkg/archive"
	"github.com/sirupsen/logrus"

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

	// forward oidc settings
	if conf.APIServer != nil && conf.APIServer.OIDC != nil {
		oidc := conf.APIServer.OIDC
		t := reflect.TypeOf(oidc).Elem()
		v := reflect.ValueOf(oidc).Elem()
		for i := 0; i < t.NumField(); i++ {
			tagValue := t.Field(i).Tag.Get("hiera")

			// skip fields without hiera tag
			if tagValue == "" {
				continue
			}

			val := v.Field(i)
			switch val.Kind() {
			case reflect.String:
				// skip empty string
				if val.String() == "" {
					continue
				}
				hieraData.variables = append(hieraData.variables, fmt.Sprintf(`%s: "%s"`, tagValue, val.String()))
			case reflect.Slice:
				// skip empty slice
				if val.Len() == 0 {
					continue
				}

				data, err := json.Marshal(val.Interface())
				if err != nil {
					panic(err)
				}
				hieraData.variables = append(hieraData.variables, fmt.Sprintf("%s: %s", tagValue, string(data)))
			default:
				panic(fmt.Sprintf("unknown type: %v", val.Kind()))
			}
		}
	}
	return
}

func kubernetesClusterConfigPerRole(conf *clusterv1alpha1.ClusterKubernetes, roleName string, hieraData *hieraData) {
	if conf == nil {
		return
	}

	if roleName == clusterv1alpha1.KubernetesMasterRoleName && conf.ClusterAutoscaler != nil && conf.ClusterAutoscaler.Enabled {
		hieraData.classes = append(hieraData.classes, `kubernetes_addons::cluster_autoscaler`)
		if conf.ClusterAutoscaler.Image != "" {
			hieraData.variables = append(hieraData.variables, fmt.Sprintf(`kubernetes_addons::cluster_autoscaler::image: "%s"`, conf.ClusterAutoscaler.Image))
		}
		if conf.ClusterAutoscaler.Version != "" {
			hieraData.variables = append(hieraData.variables, fmt.Sprintf(`kubernetes_addons::cluster_autoscaler::version: "%s"`, conf.ClusterAutoscaler.Version))
		}
	}

	if roleName == clusterv1alpha1.KubernetesMasterRoleName && conf.Tiller != nil && conf.Tiller.Enabled {
		hieraData.classes = append(hieraData.classes, `kubernetes_addons::tiller`)
		if conf.Tiller.Image != "" {
			hieraData.variables = append(hieraData.variables, fmt.Sprintf(`kubernetes_addons::tiller::image: "%s"`, conf.Tiller.Image))
		}
		if conf.Tiller.Version != "" {
			hieraData.variables = append(hieraData.variables, fmt.Sprintf(`kubernetes_addons::tiller::version: "%s"`, conf.Tiller.Version))
		}
	}

	if roleName == clusterv1alpha1.KubernetesMasterRoleName && conf.Dashboard != nil && conf.Dashboard.Enabled {
		hieraData.classes = append(hieraData.classes, `kubernetes_addons::dashboard`)
		if conf.Dashboard.Image != "" {
			hieraData.variables = append(hieraData.variables, fmt.Sprintf(`kubernetes_addons::dashboard::image: "%s"`, conf.Dashboard.Image))
		}
		if conf.Dashboard.Version != "" {
			hieraData.variables = append(hieraData.variables, fmt.Sprintf(`kubernetes_addons::dashboard::version: "%s"`, conf.Dashboard.Version))
		}
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

func contentClusterConfig(cluster interfaces.Cluster) (lines []string) {

	hieraData := &hieraData{}
	if publicAPIHostname := cluster.PublicAPIHostname(); publicAPIHostname != "" {
		sans := []string{publicAPIHostname}
		sansJSON, err := json.Marshal(&sans)
		if err != nil {
			panic(err)
		}
		hieraData.variables = append(hieraData.variables, fmt.Sprintf("tarmak::master::apiserver_additional_san_domains: %s", string(sansJSON)))
	}
	kubernetesClusterConfig(cluster.Config().Kubernetes, hieraData)

	classes, variables := serialiseHieraData(hieraData)

	return append(classes, variables...)
}

func contentInstancePoolConfig(clusterConf *clusterv1alpha1.Cluster, instanceConf *clusterv1alpha1.InstancePool, roleName string) (classes, variables []string) {

	hieraData := &hieraData{}
	kubernetesClusterConfigPerRole(clusterConf.Kubernetes, roleName, hieraData)
	kubernetesInstancePoolConfig(instanceConf.Kubernetes, hieraData)

	return serialiseHieraData(hieraData)
}

func serialiseHieraData(hieraData *hieraData) (classes, variables []string) {

	if hieraData == nil {
		return classes, variables
	}

	if len(hieraData.classes) > 0 {
		classes = append(classes, `classes:`)
		for _, class := range hieraData.classes {
			classes = append(classes, fmt.Sprintf(`- %s`, class))
		}
		classes = append(classes, "", "")
	}

	if len(hieraData.variables) > 0 {
		for _, variable := range hieraData.variables {
			variables = append(variables, fmt.Sprintf(`%s`, variable))
		}
		variables = append(variables, "", "")
	}

	return classes, variables
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
		contentClusterConfig(cluster),
	)
	if err != nil {
		return fmt.Errorf("error writing global hiera config: %s", err)
	}

	// loop through instance pools
	for _, instancePool := range cluster.InstancePools() {

		classes, variables := contentInstancePoolConfig(cluster.Config(), instancePool.Config(), instancePool.Role().Name())

		//  classes
		err = p.writeLines(
			filepath.Join(hieraPath, "instance_pools", fmt.Sprintf("%s_classes.yaml", instancePool.Name())), classes,
		)
		//  variables
		err = p.writeLines(
			filepath.Join(hieraPath, "instance_pools", fmt.Sprintf("%s_variables.yaml", instancePool.Name())), variables,
		)
		if err != nil {
			return fmt.Errorf("error writing global hiera for instancePool %s: %s", instancePool.Name(), err)
		}
	}

	return nil

}
