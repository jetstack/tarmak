# Class: vault_server
# ===========================
#
# Puppet module to install and manage a vault server install
#
# === Parameters
#
# [*version*]
#   The package version to install
#
# [*token*]
#   Static token for the vault server
#   Either token or init_token needs to be specified
#
# [*init_token*]
#   Initial token for the vault server to generate node unique token
#   Either token or init_token needs to be specified
#
# [*init_policies*]
#   TODO
#
# [*init_role*]
#   TODO
class vault_server (
  $version = $::vault_server::params::version,
  $bin_dir = $::vault_server::params::bin_dir,
  $local_bin_dir = $::vault_server::params::local_bin_dir,
  $download_dir = $::vault_server::params::download_dir,
  $dest_dir = $::vault_server::params::dest_dir,
  $server_url = $::vault_server::params::server_url,
  $systemd_dir = $::vault_server::params::systemd_dir,
  $init_token = undef,
  $init_role = undef,
  $token = undef,
  $ca_cert_path = undef,
) inherits ::vault_server::params {

  # verify inputs

  ## only one of init_token or token needs to exist
  if $init_token == undef and $token == undef {
    fail('You must provide at least one of $init_token or $token.')
  }
  if $init_token != undef and $token != undef {
    fail('You must provide either $init_token or $token.')
  }

  # paths
  $path = defined('$::path') ? {
    default => '/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/opt/bin',
    true    => $::path,
  }

  ## build download URL
  $download_url = regsubst(
    $::vault_server::params::download_url,
    '#VERSION#',
    $version,
    'G'
  )

  # token path
  $config_path = "${::vault_server::config_dir}/config"

  # token path
  $token_path = "${::vault_server::config_dir}/token"

  # init_token path
  $init_token_path = "${::vault_server::config_dir}/init-token"

  $_dest_dir = "${dest_dir}/${::vault_server::params::app_name}-${version}"

  user { 'vault':
    ensure => 'present',
    system => true,
    home   => '/var/lib/vault',
  }

  class { '::vault_server::install': }
  -> class { '::vault_server::config': }
  #-> class { '::vault_server::service': }
  -> Class['::vault_server']
}
