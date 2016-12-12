# == Class vault_client::install
#
# This class is called from vault_client for install.
#
class vault_client::install {
  $vault_bin = "${::vault_client::_dest_dir}/vault"

  file { $::vault_client::_dest_dir:
    ensure => directory,
    before => File[$vault_bin],
  }

  package { 'epel-release': ensure => present} ->
  package { 'jq': ensure => present }

  package { 'unzip': ensure => present} ->

  archive { "${::vault_client::download_dir}/vault.zip":
    ensure       => present,
    extract      => true,
    extract_path => $::vault_client::_dest_dir,
    source       => $::vault_client::download_url,
    cleanup      => true,
    creates      => $vault_bin,
  } ->

  file { $vault_bin:
    ensure => file,
    mode   => '0755',
  }

  file { '/usr/bin/vault':
    ensure => link,
    target => $vault_bin,
  }
}
