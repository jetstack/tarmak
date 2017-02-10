define vault_client::secret_service (
  String $secret_path,
  String $field,
  String $dest_path,
  String $user = 'root',
  String $group = 'root',
  Array $exec_post = [],
)
{
  $service_name = "${name}-secret"

  file { "${::vault_client::systemd_dir}/${service_name}.service":
    ensure  => file,
    content => template('vault_client/secret.service.erb'),
    notify  => Service["${service_name}.service"],
  } ~>
  exec { "${service_name}-systemctl-daemon-reload":
    command     => 'systemctl daemon-reload',
    refreshonly => true,
    path        => $::vault_client::path,
  } ->
  service { "${service_name}.service":
    ensure => 'running',
    enable => true,
  }
}
