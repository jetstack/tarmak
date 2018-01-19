class prometheus(
  String $systemd_path = '/etc/systemd/system',
  String $namespace = 'monitoring',
  Optional[Enum['etcd','master','worker']] $role = $::prometheus::params::role,
  $etcd_cluster = $::prometheus::params::etcd_cluster,
  Integer[1025,65535] $etcd_k8s_main_port = $::prometheus::params::etcd_k8s_main_port,
  Integer[1025,65535] $etcd_k8s_events_port = $::prometheus::params::etcd_k8s_events_port,
  Integer[1024,65535] $etcd_overlay_port = $::prometheus::params::etcd_overlay_port,
) inherits ::prometheus::params
{

  if $role == 'master' {
    include ::prometheus::server
    include ::prometheus::kube_state_metrics
    include ::prometheus::node_exporter
    include ::prometheus::blackbox_exporter
    include ::prometheus::blackbox_exporter_etcd
  }

  if $role == 'etcd' {
    include ::prometheus::node_exporter
    include ::prometheus::blackbox_exporter_etcd
  }

}
