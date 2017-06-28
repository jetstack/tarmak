class prometheus::prometheus_deployment (
  $addon_dir = $::prometheus::addon_dir,
  $helper_dir = $::prometheus::helper_dir,
  $prometheus_namespace = $::prometheus::prometheus_namespace,
  $prometheus_image = $::prometheus::prometheus_image,
  $prometheus_version = $::prometheus::prometheus_version,
  $prometheus_storage_local_retention = $::prometheus::prometheus_storage_local_retention,
  $prometheus_storage_local_memchunks = $::prometheus::prometheus_storage_local_memchunks,
  $prometheus_port = $::prometheus::prometheus_port,
  $prometheus_use_module_config = $::prometheus::prometheus_use_module_config,
  $etcd_cluster = $::prometheus::etcd_cluster,
  $etcd_k8s_port = $::prometheus::etcd_k8s_port,
  $etcd_events_port = $::prometheus::etcd_events_port,
  $etcd_overlay_port = $::prometheus::etcd_overlay_port,
  $prometheus_use_module_rules = $::prometheus::prometheus_use_module_rules,
  $prometheus_install_state_metrics = $::prometheus::prometheus_install_state_metrics,
  $prometheus_install_node_exporter = $::prometheus::prometheus_install_node_exporter,
  $node_exporter_image = $::prometheus::node_exporter_image,
  $node_exporter_port = $::prometheus::node_exporter_port,
  $node_exporter_version = $::prometheus::node_exporter_version,
)
{
  require ::kubernetes

  kubernetes::apply{'prometheus-server':
    manifests => [
      template('prometheus/prometheus-ns.yaml.erb'),
      template('prometheus/prometheus-config.yaml.erb'),
      template('prometheus/prometheus-rules.yaml.erb'),
      template('prometheus/prometheus-deployment.yaml.erb'),
      template('prometheus/prometheus-svc.yaml.erb'),
    ],
  }

  kubernetes::apply{'kube-state-metrics':
    manifests => [
      template('prometheus/prometheus-ns.yaml.erb'),
      template('prometheus/kube-state-metrics-deployment.yaml.erb'),
      template('prometheus/kube-state-metrics-service.yaml.erb'),
    ],
  }

  kubernetes::apply{'node-exporter':
    manifests => [
      template('prometheus/prometheus-ns.yaml.erb'),
      template('prometheus/prometheus-node-exporter-ds.yaml.erb'),
    ],
  }
}
