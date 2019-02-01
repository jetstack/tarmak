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
	"sort"
	"strings"

	"github.com/docker/docker/pkg/archive"
	multierror "github.com/hashicorp/go-multierror"
	"github.com/sirupsen/logrus"

	clusterv1alpha1 "github.com/jetstack/tarmak/pkg/apis/cluster/v1alpha1"
	tarmakv1alpha1 "github.com/jetstack/tarmak/pkg/apis/tarmak/v1alpha1"
	"github.com/jetstack/tarmak/pkg/tarmak/interfaces"
	"github.com/jetstack/tarmak/pkg/tarmak/utils"
)

type Puppet struct {
	log         *logrus.Entry
	tarmak      interfaces.Tarmak
	packerBuild bool
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

	err = p.Initialize(false)
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

func (p *Puppet) Initialize(packerBuild bool) error {
	p.packerBuild = packerBuild
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

	return nil
}

func kubernetesClusterConfig(conf *clusterv1alpha1.ClusterKubernetes, hieraData *hieraData) {
	if conf == nil {
		return
	}
	if conf.Version != "" {
		hieraData.variables = append(hieraData.variables, fmt.Sprintf(`tarmak::kubernetes_version: "%s"`, conf.Version))
	}

	// forward pod CIDR settings
	if conf.PodCIDR != "" {
		hieraData.variables = append(hieraData.variables, fmt.Sprintf(`tarmak::kubernetes_pod_network: "%s"`, conf.PodCIDR))
	}

	// forward service IP settings
	if conf.ServiceCIDR != "" {
		if parts := strings.Split(conf.ServiceCIDR, "/"); len(parts) == 2 {
			hieraData.variables = append(hieraData.variables, fmt.Sprintf(`kubernetes::service_ip_range_network: "%s"`, parts[0]))
			hieraData.variables = append(hieraData.variables, fmt.Sprintf(`kubernetes::service_ip_range_mask: "%s"`, parts[1]))
		}
	}

	// forward oidc settings
	if a := conf.APIServer; a != nil {

		if len(a.EnableAdmissionControllers) > 0 {
			hieraData.variables = append(hieraData.variables, fmt.Sprintf(`kubernetes::apiserver::admission_control: ["%s"]`, strings.Join(a.EnableAdmissionControllers, ",")))
		}

		if len(a.DisableAdmissionControllers) > 0 {
			hieraData.variables = append(hieraData.variables, fmt.Sprintf(`kubernetes::apiserver::disable_admission_control: ["%s"]`, strings.Join(a.DisableAdmissionControllers, ",")))
		}

		if a.OIDC != nil {
			oidc := a.OIDC
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
	}

	if conf.PodSecurityPolicy != nil {
		if conf.PodSecurityPolicy.Enabled {
			hieraData.variables = append(hieraData.variables, fmt.Sprintf(`tarmak::kubernetes_pod_security_policy: true`))
		} else {
			hieraData.variables = append(hieraData.variables, fmt.Sprintf(`tarmak::kubernetes_pod_security_policy: false`))
		}
	}

	// enable prometheus if set, default: enabled
	if conf.Prometheus == nil || conf.Prometheus.Enabled {
		mode := clusterv1alpha1.PrometheusModeFull
		if conf.Prometheus != nil && conf.Prometheus.Mode != "" {
			mode = conf.Prometheus.Mode
		}
		hieraData.variables = append(hieraData.variables, fmt.Sprintf("prometheus::mode: %s", mode))
		hieraData.classes = append(hieraData.classes, `prometheus`)
	}

	globalGates := make(map[string]bool)
	if conf.GlobalFeatureGates != nil {
		globalGates = conf.GlobalFeatureGates
	}

	compGates := make(map[string]bool)
	if a := conf.APIServer; a != nil {
		compGates = a.FeatureGates
	}
	if gates := featureGatesString(globalGates, compGates, true, conf.ClusterAutoscaler); gates != "" {
		hieraData.variables = append(hieraData.variables, fmt.Sprintf(`kubernetes::apiserver::feature_gates:%s`, gates))
	}

	compGates = make(map[string]bool)
	if k := conf.Kubelet; k != nil {
		compGates = k.FeatureGates
	}
	if gates := featureGatesString(globalGates, compGates, true, conf.ClusterAutoscaler); gates != "" {
		hieraData.variables = append(hieraData.variables, fmt.Sprintf(`kubernetes::kubelet::feature_gates:%s`, gates))
	}

	compGates = make(map[string]bool)
	if s := conf.Scheduler; s != nil {
		compGates = s.FeatureGates
	}
	if gates := featureGatesString(globalGates, compGates, true, conf.ClusterAutoscaler); gates != "" {
		hieraData.variables = append(hieraData.variables, fmt.Sprintf(`kubernetes::scheduler::feature_gates:%s`, gates))
	}

	compGates = make(map[string]bool)
	if p := conf.Proxy; p != nil {
		compGates = p.FeatureGates
	}
	if gates := featureGatesString(globalGates, compGates, false, conf.ClusterAutoscaler); gates != "" {
		hieraData.variables = append(hieraData.variables, fmt.Sprintf(`kubernetes::proxy::feature_gates:%s`, gates))
	}

	compGates = make(map[string]bool)
	if c := conf.ControllerManager; c != nil {
		compGates = c.FeatureGates
	}
	if gates := featureGatesString(globalGates, compGates, false, conf.ClusterAutoscaler); gates != "" {
		hieraData.variables = append(hieraData.variables, fmt.Sprintf(`kubernetes::controller_manager::feature_gates:%s`, gates))
	}

	if conf.Calico != nil {
		hieraData.variables = append(hieraData.variables, fmt.Sprintf("calico::backend: %s", conf.Calico.Backend))
		if conf.Calico.Backend == "kubernetes" {
			hieraData.variables = append(hieraData.variables, "kubernetes::controller_manager::allocate_node_cidrs: true")
		}
	}

	return
}

func featureGatesString(globalGates, componentGates map[string]bool, usePodPriority bool, conf *clusterv1alpha1.ClusterKubernetesClusterAutoscaler) string {
	gates := utils.DuplicateMapBool(globalGates)
	if usePodPriority {
		if conf != nil && conf.Overprovisioning != nil && conf.Overprovisioning.Enabled {
			gates["PodPriority"] = true
		}
	}

	if componentGates != nil {
		gates = utils.MergeMapsBool(gates, componentGates)
	}

	if len(gates) == 0 {
		return ""
	}

	var args []string
	for gate, value := range gates {
		args = append(args, fmt.Sprintf("  %s: %t", gate, value))
	}
	sort.Strings(args)

	return fmt.Sprintf("\n%s", strings.Join(args, "\n"))
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

		if conf.ClusterAutoscaler.ScaleDownUtilizationThreshold != nil {
			hieraData.variables = append(hieraData.variables, fmt.Sprintf(`kubernetes_addons::cluster_autoscaler::scale_down_utilization_threshold: %v`, *conf.ClusterAutoscaler.ScaleDownUtilizationThreshold))
		}

		if conf.ClusterAutoscaler.Overprovisioning != nil && conf.ClusterAutoscaler.Overprovisioning.Enabled {
			hieraData.variables = append(hieraData.variables, `kubernetes_addons::cluster_autoscaler::enable_overprovisioning: true`)
			hieraData.variables = append(hieraData.variables, `kubernetes::enable_pod_priority: true`)

			if conf.ClusterAutoscaler.Overprovisioning.Image != "" {
				hieraData.variables = append(hieraData.variables, fmt.Sprintf(`kubernetes_addons::cluster_autoscaler::proportional_image: "%s"`, conf.ClusterAutoscaler.Overprovisioning.Image))
			}
			if conf.ClusterAutoscaler.Overprovisioning.Version != "" {
				hieraData.variables = append(hieraData.variables, fmt.Sprintf(`kubernetes_addons::cluster_autoscaler::proportional_version: "%s"`, conf.ClusterAutoscaler.Overprovisioning.Version))
			}

			hieraData.variables = append(hieraData.variables, fmt.Sprintf(`kubernetes_addons::cluster_autoscaler::reserved_millicores_per_replica: %d`, conf.ClusterAutoscaler.Overprovisioning.ReservedMillicoresPerReplica))
			hieraData.variables = append(hieraData.variables, fmt.Sprintf(`kubernetes_addons::cluster_autoscaler::reserved_megabytes_per_replica: %d`, conf.ClusterAutoscaler.Overprovisioning.ReservedMegabytesPerReplica))
			hieraData.variables = append(hieraData.variables, fmt.Sprintf(`kubernetes_addons::cluster_autoscaler::cores_per_replica: %d`, conf.ClusterAutoscaler.Overprovisioning.CoresPerReplica))
			hieraData.variables = append(hieraData.variables, fmt.Sprintf(`kubernetes_addons::cluster_autoscaler::nodes_per_replica: %d`, conf.ClusterAutoscaler.Overprovisioning.NodesPerReplica))
			hieraData.variables = append(hieraData.variables, fmt.Sprintf(`kubernetes_addons::cluster_autoscaler::replica_count: %d`, conf.ClusterAutoscaler.Overprovisioning.ReplicaCount))
		}
	}

	if roleName == clusterv1alpha1.KubernetesWorkerRoleName && conf.ClusterAutoscaler != nil && conf.ClusterAutoscaler.Enabled {
		if conf.ClusterAutoscaler.Overprovisioning != nil && conf.ClusterAutoscaler.Overprovisioning.Enabled {
			hieraData.variables = append(hieraData.variables, `kubernetes::enable_pod_priority: true`)
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

func (p *Puppet) contentClusterConfig(cluster interfaces.Cluster) ([]string, error) {

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

	hieraData.classes = append(hieraData.classes, `tarmak::fluent_bit`)
	if cluster.Config().LoggingSinks != nil && len(cluster.Config().LoggingSinks) > 0 {
		jsonLoggingSink, err := json.Marshal(cluster.Config().LoggingSinks)
		if err != nil {
			return nil, fmt.Errorf("unable to marshall logging sinks: %s", err)
		}
		hieraData.variables = append(hieraData.variables, fmt.Sprintf(`tarmak::fluent_bit_configs: %s`, string(jsonLoggingSink)))
	}

	if v := cluster.Config().VaultHelper; v != nil {
		if v.URL != "" {
			hieraData.variables = append(hieraData.variables, fmt.Sprintf("vault_client::_download_url: %s", v.URL))
		}
	}

	if p.packerBuild {
		hieraData.variables = append(hieraData.variables, `vault_client::run_exec: false`)
		hieraData.variables = append(hieraData.variables, `tarmak::service_ensure: "stopped"`)
	}

	classes, variables := serialiseHieraData(hieraData)

	return append(classes, variables...), nil
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

	clusterConfig, err := p.contentClusterConfig(cluster)
	if err != nil {
		return fmt.Errorf("unable to retrieve cluster config: %s", err)
	}
	// write cluster config
	err = p.writeLines(
		filepath.Join(hieraPath, "tarmak.yaml"),
		clusterConfig,
	)
	if err != nil {
		return fmt.Errorf("error writing global hiera config: %s", err)
	}

	// build instance pool arrays
	workerMinCounts := make([]int, 0)
	workerMaxCounts := make([]int, 0)
	workerInstancePoolNames := make([]string, 0)
	if cluster.Config().Kubernetes.ClusterAutoscaler != nil && cluster.Config().Kubernetes.ClusterAutoscaler.Enabled {
		for _, instancePool := range cluster.InstancePools() {
			if instancePool.Role().Name() == clusterv1alpha1.KubernetesWorkerRoleName {
				workerMinCounts = append(workerMinCounts, instancePool.MinCount())
				workerMaxCounts = append(workerMaxCounts, instancePool.MaxCount())
				workerInstancePoolNames = append(workerInstancePoolNames, instancePool.Name())
			}
		}
	}

	// loop through instance pools
	for _, instancePool := range cluster.InstancePools() {

		classes, variables := contentInstancePoolConfig(cluster.Config(), instancePool.Config(), instancePool.Role().Name())

		if instancePool.Role().Name() == clusterv1alpha1.KubernetesMasterRoleName && cluster.Config().Kubernetes.ClusterAutoscaler != nil && cluster.Config().Kubernetes.ClusterAutoscaler.Enabled {
			s, err := json.Marshal(workerMinCounts)
			if err != nil {
				panic(err)
			}
			variables = append(variables, fmt.Sprintf(`kubernetes_addons::cluster_autoscaler::min_instances: %s`, string(s)))
			s, err = json.Marshal(workerMaxCounts)
			if err != nil {
				panic(err)
			}
			variables = append(variables, fmt.Sprintf(`kubernetes_addons::cluster_autoscaler::max_instances: %s`, string(s)))
			variables = append(variables, fmt.Sprintf(`kubernetes_addons::cluster_autoscaler::instance_pool_names: ['%s']`, strings.Join(workerInstancePoolNames[:], `','`)))
		}

		var taintLabelError error

		if len(instancePool.Config().Labels) > 0 {
			labels, err := instancePool.Labels()
			if err != nil {
				taintLabelError = multierror.Append(taintLabelError, fmt.Errorf("error reading instance pool labels: %s", err))
			} else {
				variables = append(variables, fmt.Sprintf("kubernetes::kubelet::node_labels:\n%s", labels))
			}
		}
		if len(instancePool.Config().Taints) > 0 {
			taints, err := instancePool.Taints()
			if err != nil {
				taintLabelError = multierror.Append(taintLabelError, fmt.Errorf("error reading instance pool taints: %s", err))
			} else {
				variables = append(variables, fmt.Sprintf("kubernetes::kubelet::node_taints:\n%s", taints))
			}
		}

		if taintLabelError != nil {
			return taintLabelError
		}

		// etcd
		if instancePool.Role().Name() == clusterv1alpha1.KubernetesEtcdRoleName {
			variables = append(variables, fmt.Sprintf(`tarmak::etcd_instances: %d`, instancePool.MinCount()))
			variables = append(variables, `tarmak::etcd_mount_unit: "var-lib-etcd.mount"`)
		}

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
