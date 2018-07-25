define vault_client::download_unsealer (
  String $user = 'root',
  String $group = 'root',
)
{
  $script_name = 'download-vault-unsealer'

  file { "${::vault_client::local_bin_dir}/${script_name}.sh":
    ensure  => file,
    content => template('vault_client/download-vault-unsealer.sh'),
    notify  => Script["${script_name}.sh"],
    owner   => $user,
    group   => $group,
    mode    => '0755',
  }
  ~> exec { "${script_name}-script-run":
    command     => "${::vault_client::local_bin_dir}/${script_name}.sh",
    refreshonly => true,
    path        => $::vault_client::path,
  }
}
