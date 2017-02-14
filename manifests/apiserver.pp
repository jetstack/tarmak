# class kubernetes::master
class kubernetes::apiserver(
  $admission_control = undef,
  $etcd_nodes = ['localhost'],
  $etcd_port = 2379,
  $etcd_events_port = undef,
  $etcd_ca_file = undef,
  $etcd_cert_file = undef,
  $etcd_key_file = undef,
  $ca_file = undef,
  $cert_file = undef,
  $key_file = undef,
  $systemd_wants = [],
  $systemd_requires = [],
  $systemd_after = [],
  $systemd_before = [],
  $insecure_bind_address = undef,
)  {
  require ::kubernetes

  $_systemd_wants = $systemd_wants
  $_systemd_requires = $systemd_requires
  $_systemd_after = ['network.target'] + $systemd_after
  $_systemd_before = $systemd_before

  # Admission controllers
  if $admission_control == undef {
    # No DefaultStorageClass controller pre 1.4
    if versioncmp($::kubernetes::version, '1.4.0') >= 0 {
      $_admission_control =  ['NamespaceLifecycle', 'LimitRanger', 'ServiceAccount', 'DefaultStorageClass', 'ResourceQuota']
    } else {
      $_admission_control =  ['NamespaceLifecycle', 'LimitRanger', 'ServiceAccount', 'ResourceQuota']

    }
  } else {
    $_admission_control = $admission_control
  }

  $service_name = 'kube-apiserver'

  if $etcd_ca_file == undef and $etcd_cert_file == undef and $etcd_key_file == undef {
    $etcd_proto = 'http'
  } else {
    $etcd_proto = 'https'
  }

  $_etcd_urls = map($etcd_nodes) |$node| { "${etcd_proto}://${node}:${etcd_port}" }
  $etcd_servers = $_etcd_urls.join(',')

  if $etcd_events_port == undef {
    $etcd_servers_overrides = []
  }
  else {
    $_etcd_events_urls = map($etcd_nodes) |$node| { "${etcd_proto}://${node}:${etcd_events_port}" }
    $etcd_events_servers = $_etcd_events_urls.join(',')
    $etcd_servers_overrides = [
      "/events#${etcd_events_servers}",
    ]
  }

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
