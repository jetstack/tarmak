class vault_client::service {
  $systemd_dir = '/etc/systemd/system'
  $service_name = $::vault_client::token_service_name
  $frequency = 86400

  if $::vault_client::ca_cert_path != undef {
    $ca_cert_path = $::vault_client::ca_cert_path
  }

  $path = defined('$::path') ? {
    default => '/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/opt/bin',
    true    => $::path
  }

  exec { "${module_name}-systemctl-daemon-reload":
    command     => 'systemctl daemon-reload',
    refreshonly => true,
    path        => $path,
  }

  file { "${systemd_dir}/${service_name}.service":
    ensure  => file,
    content => template('vault_client/token-renewal.service.erb'),
    notify  => Exec["${module_name}-systemctl-daemon-reload"],
  } ~>
  exec { "${service_name}-trigger":
    command     => "systemctl start ${service_name}.service",
    path        => $path,
    refreshonly => true,
    require     => Exec["${module_name}-systemctl-daemon-reload"],
  }

  file { "${systemd_dir}/${service_name}.timer":
    ensure  => file,
    content => template('vault_client/token-renewal.timer.erb'),
    notify  => Exec["${module_name}-systemctl-daemon-reload"],
  } ~>
  service { "${service_name}.timer":
    ensure => 'running',
    enable => true,
  }

}
