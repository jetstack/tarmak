class vault_server::service (
  String $region = $vault_server::region,
  String $vault_unsealer_kms_key_id = $vault_server::vault_unsealer_kms_key_id,
  String $vault_unsealer_ssm_key_prefix = $vault_server::vault_unsealer_ssm_key_prefix,
  String $user = 'root',
  String $group = 'root',
  String $assets_service_name = 'vault-assets',
  String $unsealer_service_name = 'vault-unsealer',
  String $service_name = 'vault',
)
{

  if $vault_server::vault_tls_cert_path == '' {
    $vault_tls_cert_path = undef
  } else {
    $vault_tls_cert_path = $vault_server::vault_tls_cert_path
  }

  if $vault_server::vault_tls_ca_path == '' {
    $vault_tls_ca_path = undef
  } else {
    $vault_tls_ca_path = $vault_server::vault_tls_ca_path
  }

  if $vault_server::vault_tls_key_path == '' {
    $vault_tls_key_path = undef
  } else {
    $vault_tls_key_path = $vault_server::vault_tls_key_path
  }

  exec { "${service_name}-systemctl-daemon-reload":
    command     => '/bin/systemctl daemon-reload',
    refreshonly => true,
    path        => defined('$::path') ? {
      default => '/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/opt/bin',
      true    => $::path
    },
  }

  file { "${::vault_server::systemd_dir}/${assets_service_name}.service":
    ensure  => file,
    content => template('vault_server/vault-assets.service.erb'),
    owner   => $user,
    group   => $group,
    mode    => '0644',
    notify  => Exec["${service_name}-systemctl-daemon-reload"]
  } ~> service { "${assets_service_name}.service":
    ensure  => 'stopped',
    enable  => false,
    require => Exec["${service_name}-systemctl-daemon-reload"],
  }

  file { "${::vault_server::systemd_dir}/${unsealer_service_name}.service":
    ensure  => file,
    content => template('vault_server/vault-unsealer.service.erb'),
    owner   => $user,
    group   => $group,
    mode    => '0644',
    notify  => Exec["${service_name}-systemctl-daemon-reload"],
  } ~> service { "${unsealer_service_name}.service":
    ensure  => 'running',
    enable  => true,
    require => Exec["${service_name}-systemctl-daemon-reload"],
  }

  file { "${::vault_server::systemd_dir}/${service_name}.service":
    ensure  => file,
    content => template('vault_server/vault.service.erb'),
    owner   => $user,
    group   => $group,
    mode    => '0644',
    notify  => Exec["${service_name}-systemctl-daemon-reload"]
  } ~> service { "${service_name}.service":
    ensure  => 'running',
    enable  => true,
    require => Exec["${service_name}-systemctl-daemon-reload"],
  }
}
