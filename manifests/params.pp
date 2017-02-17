class prometheus::params {
  $blackbox_download_url = 'https://github.com/jetstack-experimental/blackbox_exporter/releases/download/poc-proxy/'
  $blackbox_dest_dir = '/usr/local/sbin'
  $blackbox_config_dir = '/etc/blackbox'
  $systemd_path = '/etc/systemd/system'
  $addon_dir = '/etc/kubernetes/addons'
  $helper_dir = '/usr/local/sbin'
  $node_exporter_image = 'prom/node-exporter'
  $node_exporter_version = 'v0.13.0'
  $node_exporter_port = 9100
  $prometheus_namespace = 'monitoring'
  $prometheus_image = 'quay.io/prometheus/prometheus'
  $prometheus_version = 'v1.4.1'
  $prometheus_port = 9090
  $prometheus_storage_local_retention = '6h'
  $prometheus_storage_local_memchunks = 500000
  $prometheus_use_module_config = true
  $prometheus_use_module_rules = true
  $prometheus_install_state_metrics = true
  $prometheus_install_node_exporter = true

  if defined('::puppernetes') {
    $etcd_cluster = $::puppernetes::_etcd_cluster
    $etcd_k8s_port = $::puppernetes::etcd_k8s_main_client_port
    $etcd_events_port = $::puppernetes::etcd_k8s_events_client_port
    $etcd_overlay_port = $::puppernetes::etcd_overlay_client_port
  } else {
    $etcd_cluster = undef
    $etcd_k8s_port = 2379
    $etcd_events_port = 2369
    $etcd_overlay_port = 2359
  }
}
