#calico params.pp
class calico::params {
  $etcd_cert_path = '/etc/etcd/ssl'
  $etcd_cert_base_name = 'etcd-overlay'
  if defined('::puppernetes') {
    $etcd_cluster = $::puppernetes::_etcd_cluster
    $etcd_overlay_port = $::puppernetes::etcd_overlay_client_port
    $bin_dir = $::puppernetes::bin_dir
    $systemd_dir = $::puppernetes::systemd_dir
    $cloud_provider = $::puppernetes::cloud_provider
  } else {
    $etcd_cluster = []
    $etcd_overlay_port = 2359
    $bin_dir = '/opt/bin'
    $systemd_dir = '/etc/systemd/system'
    $cloud_provider = ''
  }
}
