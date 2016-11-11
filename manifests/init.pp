# Class: vault_client
# ===========================
#
# Full description of class vault_client here.
#
class vault_client (
  $version = $::vault_client::params::version,
  $bin_dir = $::vault_client::params::bin_dir,
  $download_dir = $::vault_client::params::download_dir,
  $dest_dir = $::vault_client::params::dest_dir,
) inherits ::vault_client::params {

  # Build download URL
  $download_url = regsubst(
    $::vault_client::params::download_url,
    '#VERSION#',
    $version,
    'G'
  )

  $_dest_dir = "${dest_dir}/${::vault_client::params::app_name}-${version}"

  # validate parameters here
  class { '::vault_client::install': } ->
  class { '::vault_client::config': } ->
  Class['::vault_client']
}
