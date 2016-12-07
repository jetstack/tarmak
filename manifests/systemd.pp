define etcd::systemd (
  String $etcd_cluster_name,
  String $etcd_version,
)
{
  file { "/usr/lib/systemd/system/etcd-${etcd_cluster_name}.service":
    ensure  => file,
    content => template('etcd/etcd.service.erb'),
    require => [ Class['etcd'], Exec["Trigger ${etcd_cluster_name} cert"] ],
  }
}
