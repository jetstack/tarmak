define vault_client::assets_service (
  String $vault_tls_cert_path,
  String $vault_tls_key_path,
  String $vault_tls_ca_path,
  String $user = 'root',
  String $group = 'root',
)
{
  $service_name = 'vault-assets'

  file { "${::vault_client::systemd_dir}/${service_name}.service":
    ensure  => file,
    content => template('vault_client/vault-assets.service.erb'),
    notify  => Service["${service_name}.service"],
    owner   => $user,
    group   => $group,
    mode    => '0644',
  }
  ~> exec { "${service_name}-systemctl-daemon-reload":
    command     => 'systemctl daemon-reload',
    refreshonly => true,
    path        => $::vault_client::path,
  }
}
