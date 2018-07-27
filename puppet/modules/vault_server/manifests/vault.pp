define vault_consul::vault (
  String $user = 'root',
  String $group = 'root',
)
{
  $script_name = 'vault'

  file { "${::vault_consul::local_bin_dir}/${script_name}.sh":
    ensure  => file,
    content => file('vault_consul/vault.sh'),
    owner   => $user,
    group   => $group,
    mode    => '0644',
  }
}
