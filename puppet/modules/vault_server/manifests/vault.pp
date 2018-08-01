define vault_server::vault (
  String $user = 'root',
  String $group = 'root',
)
{
  $script_name = 'vault'

  file { "${::vault_server::local_bin_dir}/${script_name}.sh":
    ensure  => file,
    content => file('vault_server/vault.sh'),
    owner   => $user,
    group   => $group,
    mode    => '0644',
  }
}
