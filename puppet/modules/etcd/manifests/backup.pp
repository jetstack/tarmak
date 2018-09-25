define etcd::backup (
  String $backup_bucket_prefix = $etcd::backup_bucket_prefix,
  String $backup_schedule = $etcd::backup_schedule,
  String $endpoints = '',
  String $ca_path = '',
){
  include ::etcd

  $service_name = "etcd-${name}"
  $backup_service_name = "${service_name}-backup"

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
