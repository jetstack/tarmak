class etcd::service (
  String $backup_bucket_prefix = $etcd::backup_bucket_prefix,
  String $backup_schedule = $etcd::backup_schedule,
  Array $initial_cluster = [],
  Boolean $tls = false,
  Integer[1,65535] $etcd_overlay_client_port = 2360,
  Integer[1,65535] $etcd_k8s_events_client_port = 2370,
  Integer[1,65535] $etcd_k8s_main_client_port = 2380,
  String $k8s_main_ca_name = '',
  String $k8s_events_ca_name = '',
  String $overlay_ca_name = '',
){
  include ::etcd

  $service_name = 'etcd'
  $backup_service_name = "${service_name}-backup"

  if $tls {
    $proto = 'https'
  } else {
    $proto = 'http'
  }

  $overlay_endpoints = "https://127.0.0.1:${etcd_overlay_client_port}"
  $k8s_events_endpoints = "https://127.0.0.1:${etcd_k8s_events_client_port}"
  $k8s_main_endpoints = "https://127.0.0.1:${etcd_k8s_main_client_port}"

  exec { "${service_name}-systemctl-daemon-reload":
    command     => 'systemctl daemon-reload',
    refreshonly => true,
    path        => defined('$::path') ? {
      default => '/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/opt/bin',
      true    => $::path
    },
  }

  file { "${etcd::systemd_dir}/${backup_service_name}.service":
    ensure  => file,
    content => template('etcd/etcd-backup.service.erb'),
    notify  => Exec["${service_name}-systemctl-daemon-reload"],
    mode    => '0644'
  }
  ~> service { "${backup_service_name}.service":
    ensure  => 'stopped',
    enable  => false,
    require => Exec["${service_name}-systemctl-daemon-reload"],
  }

  file { "${etcd::systemd_dir}/${backup_service_name}.timer":
    ensure  => file,
    content => template('etcd/etcd-backup.timer.erb'),
    notify  => Exec["${service_name}-systemctl-daemon-reload"],
  }
  ~> service { "${backup_service_name}.timer":
    ensure  => 'running',
    enable  => true,
    require => Exec["${service_name}-systemctl-daemon-reload"],
  }
}
