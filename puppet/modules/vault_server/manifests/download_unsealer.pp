define vault_server::download_unsealer (
  String $user = 'root',
  String $group = 'root',
)
{
  $script_name = 'download-vault-unsealer'

  file { "${::vault_server::local_bin_dir}/${script_name}.sh":
    ensure  => file,
    content => file('vault_server/download-vault-unsealer.sh'),
    owner   => $user,
    group   => $group,
    mode    => '0755',
  }
  ~> exec { "${script_name}-script-run":
    command => "${::vault_server::local_bin_dir}/${script_name}.sh",
    path    => $::vault_server::path,
  }
}
