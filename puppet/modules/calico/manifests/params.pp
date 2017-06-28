#calico params.pp
class calico::params {
  $etcd_cert_path = '/etc/etcd/ssl'
  $etcd_cert_base_name = 'etcd-overlay'
  if defined('::tarmak') {
    $etcd_cluster = $::tarmak::_etcd_cluster
    $etcd_overlay_port = $::tarmak::etcd_overlay_client_port
    $bin_dir = $::tarmak::bin_dir
    $systemd_dir = $::tarmak::systemd_dir
    $cloud_provider = $::tarmak::cloud_provider
  } else {
    $etcd_cluster = []
    $etcd_overlay_port = 2359
    $bin_dir = '/opt/bin'
    $systemd_dir = '/etc/systemd/system'
    $cloud_provider = ''
  }
}
