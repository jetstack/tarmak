class prometheus(
  String $systemd_path = '/etc/systemd/system',
  String $namespace = 'monitoring',
  Optional[Enum['etcd','master','worker', 'vault']] $role = $::prometheus::params::role,
  $etcd_cluster_exporters = $::prometheus::params::etcd_cluster_exporters,
  Optional[Integer[1025,65535]] $etcd_k8s_main_port = $::prometheus::params::etcd_k8s_main_port,
  Optional[Integer[1025,65535]] $etcd_k8s_events_port = $::prometheus::params::etcd_k8s_events_port,
  Optional[Integer[1024,65535]] $etcd_overlay_port = $::prometheus::params::etcd_overlay_port,
  String $mode = 'Full',
  Enum['present', 'absent'] $ensure = 'present',
) inherits ::prometheus::params
{
  if $ensure == 'present' {
    $service_ensure = 'running'
    $service_enable = true
  } else {
    $service_ensure = 'stopped'
    $service_enable = false
  }

  if $role == 'master' {
    if $mode == 'Full' {
      include ::prometheus::server
      include ::prometheus::blackbox_exporter_etcd
      include ::prometheus::node_exporter

      include ::prometheus::kube_state_metrics
      include ::prometheus::blackbox_exporter
    }

    if $mode == 'ExternalScrapeTargetsOnly' {
      include ::prometheus::server
      include ::prometheus::blackbox_exporter_etcd
      include ::prometheus::node_exporter
    }
  }

  if $role == 'etcd' {
    include ::prometheus::node_exporter
    include ::prometheus::blackbox_exporter_etcd
  }

}
