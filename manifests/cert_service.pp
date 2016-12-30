define vault_client::cert_service (
  String $base_path,
  String $common_name,
  String $role,
  String $extra_opts = '',
  String $key_type = 'rsa',
  Integer $key_bits = 2048,
  Integer $frequency = 86400,
  String $user = 'root',
  String $group = 'root',
  Array $exec_post = [],
)
{
  $systemd_dir = '/etc/systemd/system'
  $service_name = "${name}-cert"

  $path = defined('$::path') ? {
    default => '/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/opt/bin',
    true    => $::path,
  }

  exec { "${service_name}-systemctl-daemon-reload":
    command     => 'systemctl daemon-reload',
    refreshonly => true,
    path        => $path,
  }

  file { "${systemd_dir}/${service_name}.service":
    ensure  => file,
    content => template('vault_client/cert.service.erb'),
    notify  => Exec["${service_name}-systemctl-daemon-reload"],
  } ~>
  exec { "${service_name}-remove-existing-certs":
    command     => "rm -rf ${base_path}-key.pem ${base_path}-csr.pem",
    path        => $path,
    refreshonly => true,
  } ~>
  exec { "${service_name}-trigger":
    command     => "systemctl start ${service_name}.service",
    path        => $path,
    refreshonly => true,
  }

  file { "${systemd_dir}/${service_name}.timer":
    ensure  => file,
    content => template('vault_client/cert.timer.erb'),
    notify  => Exec["${service_name}-systemctl-daemon-reload"],
  } ~>
  service { "${service_name}.timer":
    ensure => 'running',
    enable => true,
  }

}
