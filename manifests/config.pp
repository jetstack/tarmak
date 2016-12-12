define calico::config (
  String $calico_name,
  Integer $etcd_count,
  Integer $calico_etcd_port,
)
{
  file { "/etc/cni/net.d/10-calico.conf":
    ensure => file,
    content => template('calico/10-calico.conf.erb'),
    require => Class['calico'],
  }
}
