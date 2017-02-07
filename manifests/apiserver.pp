# class kubernetes::master
class kubernetes::apiserver(
  $admission_control = ['NamespaceLifecycle', 'LimitRanger', 'ServiceAccount', 'DefaultStorageClass', 'ResourceQuota'],
  $etcd_nodes = ['127.0.0.1'],
  $etcd_port = 2379,
  $etcd_events_port = undef,
  $etcd_ca_file = undef,
  $etcd_cert_file = undef,
  $etcd_key_file = undef,
  $ca_file = undef,
  $cert_file = undef,
  $key_file = undef,
)  {
  include ::kubernetes

  $service_name = 'kube-apiserver'

  if $etcd_ca_file == undef and $etcd_cert_file == undef and $etcd_key_file == undef {
    $etcd_proto = 'http'
  } else {
    $etcd_proto = 'https'
  }

  $_etcd_urls = map($etcd_nodes) |$node| { "${etcd_proto}://${node}:${etcd_port}" }
  $etcd_servers = $_etcd_urls.join(',')
  $etcd_servers_overrides = []

  kubernetes::symlink{'apiserver':} ->
  file{"${::kubernetes::systemd_dir}/${service_name}.service":
    ensure  => file,
    mode    => '0640',
    owner   => 'root',
    group   => 'root',
    content => template("kubernetes/${service_name}.service.erb"),
    notify  => Service["${service_name}.service"],
  } ~>
  exec { "${service_name}-daemon-reload":
    command     => 'systemctl daemon-reload',
    path        => $::kubernetes::path,
    refreshonly => true,
  } ->
  service{ "${service_name}.service":
    ensure => running,
    enable => true,
  }

}
