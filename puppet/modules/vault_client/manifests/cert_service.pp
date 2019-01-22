include ::tarmak
define vault_client::cert_service (
  String $base_path,
  String $common_name,
  String $role,
  Array[String] $alt_names = [],
  Array[String] $ip_sans = [],
  Integer $uid = 0,
  Integer $gid = 0,
  String $key_type = 'rsa',
  Integer $key_bits = 2048,
  Integer $frequency = 86400,
  Array $exec_post = [],
  Boolean $run_exec = true,
)
{
  require vault_client

  $service_name = "${name}-cert"
  exec { "${service_name}-systemctl-daemon-reload":
    command     => 'systemctl daemon-reload',
    refreshonly => true,
    path        => $::vault_client::path,
  }

  file { "${::vault_client::systemd_dir}/${service_name}.service":
    ensure  => file,
    content => template('vault_client/cert.service.erb'),
    notify  => Exec["${service_name}-systemctl-daemon-reload", "${service_name}-trigger"],
  }
  ~> exec { "${service_name}-remove-existing-certs":
    command     => "rm -rf ${base_path}-key.pem ${base_path}-csr.pem",
    path        => $::vault_client::path,
    refreshonly => true,
    require     => Exec["${service_name}-systemctl-daemon-reload"],
  }
  ~> service { "${service_name}.service":
    enable  => true,
    require => Exec["${service_name}-systemctl-daemon-reload"],
  }

  if $run_exec {
    $trigger_cmd = "systemctl start ${service_name}.service"
  } else {
    $trigger_cmd = '/bin/true'
  }
  exec { "${service_name}-trigger":
    path        => $::vault_client::path,
    command     => $trigger_cmd,
    refreshonly => true,
    require     => Exec["${service_name}-systemctl-daemon-reload"],
  }
  -> exec { "${service_name}-create-if-missing":
    command => $trigger_cmd,
    creates => "${base_path}.pem",
    path    => $::vault_client::path,
    require => Exec["${service_name}-systemctl-daemon-reload"],
  }

  file { "${vault_client::systemd_dir}/${service_name}.timer":
    ensure  => file,
    content => template('vault_client/cert.timer.erb'),
    notify  => Exec["${service_name}-systemctl-daemon-reload"],
  }
  ~> service { "${service_name}.timer":
    ensure  => 'running',
    enable  => true,
    require => Exec["${service_name}-systemctl-daemon-reload"],
  }
}
