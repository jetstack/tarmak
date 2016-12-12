# Class: vault_client
# ===========================
#
# Puppet module to install and manage a vault client install
#
# === Parameters
#
# [*version*]
#   The package version to install
#
# [*token*]
#   Static token for the vault client
#   Either token or init_token needs to be specified
#
# [*init_token*]
#   Initial token for the vault client to generate node unique token
#   Either token or init_token needs to be specified
#
# [*init_policies*]
#   TODO
#
# [*init_role*]
#   TODO
class vault_client (
  $version = $::vault_client::params::version,
  $bin_dir = $::vault_client::params::bin_dir,
  $download_dir = $::vault_client::params::download_dir,
  $dest_dir = $::vault_client::params::dest_dir,
  $server_url = $::vault_client::params::server_url,
  $init_token = undef,
  $init_policies = [],
  $init_role = undef,
  $token = undef,
) inherits ::vault_client::params {

  # verify inputs

  ## only one of init_token or token needs to exist
  if $init_token == undef and $token == undef {
    fail('You must provide at least one of $init_token or $token.')
  }
  if $init_token != undef and $token != undef {
    fail('You must provide either $init_token or $token.')
  }

  # paths

  ## build download URL
  $download_url = regsubst(
    $::vault_client::params::download_url,
    '#VERSION#',
    $version,
    'G'
  )

  # token path
  $config_path = "${::vault_client::config_dir}/config"

  # helper script path
  $helper_path = "${::vault_client::config_dir}/helper"

  # token path
  $token_path = "${::vault_client::config_dir}/token"

  # init_token path
  $init_token_path = "${::vault_client::config_dir}/init-token"

  # token renewal service
  $token_service_name = 'vault-token-renewal'

  $_dest_dir = "${dest_dir}/${::vault_client::params::app_name}-${version}"

  class { '::vault_client::install': } ->
  class { '::vault_client::config': } ->
  class { '::vault_client::service': } ->
  Class['::vault_client']
}
