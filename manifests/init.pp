class prometheus(
  $role = '',
  $etcd_cluster = '',
  $etcd_k8s_port = $::prometheus::params::etcd_k8s_port,
  $etcd_events_port = $::prometheus::params::etcd_events_port,
  $etcd_overlay_port = $::prometheus::params::etcd_overlay_port,
  $blackbox_download_url = $::prometheus::params::blackbox_download_url,
  $blackbox_dest_dir = $::prometheus::params::blackbox_dest_dir,
  $systemd_path = $::prometheus::params::systemd_path,
  $node_exporter_image = $::prometheus::params::node_exporter_image,
  $node_exporter_version = $::prometheus::params::node_exporter_version,
  $node_exporter_port = $::prometheus::params::node_exporter_port,
  $addon_dir = $::prometheus::params::addon_dir,
  $helper_dir = $::prometheus::params::helper_dir,
  $prometheus_namespace = $::prometheus::params::prometheus_namespace,
  $prometheus_image = $::prometheus::params::prometheus_image,
  $prometheus_version = $::prometheus::params::prometheus_version,
  $prometheus_storage_local_retention = $::prometheus::params::prometheus_retention,
  $prometheus_storage_local_memchunks = $::prometheus::params::prometheus_storage_local_memchunks,
  $prometheus_port = $::prometheus::params::prometheus_port,
  $prometheus_use_module_config = $::prometheus::params::prometheus_use_module_config,
  $prometheus_use_module_rules = $::prometheus::params::prometheus_use_module_rules,
  $prometheus_install_state_metrics = $::prometheus::params::prometheus_install_state_metrics,
  $prometheus_install_node_exporter = $::prometheus::params::prometheus_install_node_exporter
) inherits prometheus::params
{
  if $role == 'etcd' {
    include prometheus::blackbox_etcd
    include prometheus::node_exporter_service
  }

  if $role == 'master' {
    include prometheus::prometheus_deployment
  }
}
