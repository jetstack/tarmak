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

  $download_url = $::vault_client::params::download_url
  $_dest_dir = "${dest_dir}/${::vault_client::params::name}-${version}"

  # validate parameters here
  class { '::vault_client::install': } ->
  class { '::vault_client::config': } ->
  Class['::vault_client']
}
