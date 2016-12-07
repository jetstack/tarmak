define etcd::config (
  String $etcd_cluster_name,
  String $etcd_version,
  Integer $client_port,
  Integer $peer_port,
)
{
  file { "/etc/etcd/etcd-${etcd_cluster_name}.conf":
    ensure  => file,
    content => template('etcd/etcd.conf.erb'),
    require => Class['etcd'],
  }
}
