define vault_client::vault_service (
  String $region,
  String $user = 'root',
  String $group = 'root',
)
{
  $service_name = 'vault'

  file { "${::vault_client::systemd_dir}/${service_name}.service":
    ensure  => file,
    content => template('vault_client/vault.service.erb'),
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
  -> service { "${service_name}.service":
    ensure => 'running',
    enable => true,
  }
}