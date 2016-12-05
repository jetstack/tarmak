define vault::etcd_cert_service (
  String $etcd_cluster,
  String $frequency,
)
{

  include ::systemd

  file { "/usr/lib/systemd/system/etcd-${etcd_cluster}-cert.service":
    ensure => file,
    content => template('etcd/etcd-cert.service.erb'),
  } ~>
  Exec['systemctl-daemon-reload']

  file { "/usr/lib/systemd/system/etcd-${etcd_cluster}-cert.timer":
    ensure => file,
    content => template('etcd/etcd-cert.service.erb'),
  } ~>
  Exec['systemctl-daemon-reload']
}


