#calico params.pp
class calico::params {
  $etcd_cert_path = '/etc/etcd/ssl'
  $etcd_cert_base_name = 'etcd-overlay'
  if defined('::tarmak') {
    $bin_dir = $::tarmak::bin_dir
    $systemd_dir = $::tarmak::systemd_dir
    $cloud_provider = $::tarmak::cloud_provider

    $etcd_cluster = $::tarmak::_etcd_cluster
    $etcd_overlay_port = $::tarmak::etcd_overlay_client_port

    $crd_feature_gates = $::tarmak::feature_gates
    $enable_typha = $::tarmak::enable_typha

  } else {
    $etcd_cluster = []
    $etcd_overlay_port = 2359
    $bin_dir = '/opt/bin'
    $systemd_dir = '/etc/systemd/system'
    $cloud_provider = ''
  }
}
