define vault_server::vault (
  String $user = 'root',
  String $group = 'root',
)
{
  $script_name = 'vault'

  file { "/etc/profile.d/${script_name}.sh":
    ensure  => file,
    content => file('vault_server/vault.sh'),
    owner   => $user,
    group   => $group,
    mode    => '0644',
  }
}
