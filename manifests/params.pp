# == Class vault_client::params
#
# This class is meant to be called from vault_client.
# It sets variables according to platform.
#
class vault_client::params {
  $app_name = 'vault-helper'
  $version = '0.8.6'
  $bin_dir = '/opt/bin'
  $dest_dir = '/opt'
  $config_dir = '/etc/vault'
  $download_dir = '/tmp'
  $download_url = 'https://github.com/jetstack-experimental/vault-helper/releases/download/#VERSION#/vault-helper_#VERSION#_linux_amd64'
  $server_url = 'http://127.0.0.1:8200'
  $systemd_dir = '/etc/systemd/system'
}
