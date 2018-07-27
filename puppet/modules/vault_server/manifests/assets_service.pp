define vault_server::assets_service (
  String $vault_tls_cert_path,
  String $vault_tls_key_path,
  String $vault_tls_ca_path,
  String $user = 'root',
  String $group = 'root',
)
{
  $service_name = 'vault-assets'

  file { "${::vault_server::systemd_dir}/${service_name}.service":
    ensure  => file,
    content => template('vault_server/vault-assets.service.erb'),
    notify  => Service["${service_name}.service"],
    owner   => $user,
    group   => $group,
    mode    => '0644',
  }
  ~> exec { "${service_name}-systemctl-daemon-reload":
    command     => 'systemctl daemon-reload',
    refreshonly => true,
    path        => $::vault_server::path,
  }
  -> service { "${service_name}.service":
    ensure => 'stopped',
    enable => false,
  }
}
