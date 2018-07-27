# == Class vault_server::params
#
# This class is meant to be called from vault_server.
# It sets variables according to platform.
#
class vault_server::params {
  $app_name = 'vault'
  $version = '0.9.5'
  $bin_dir = '/opt/bin'
  $local_bin_dir = '/usr/local/bin'
  $dest_dir = '/opt'
  $config_dir = '/etc/vault'
  $download_dir = '/tmp'
  #$download_url = "https://releases.hashicorp.com/vault/#VERSION#/vault_#VERSION_linux_amd64.zip"
  #$server_url = 'http://127.0.0.1:8200'
  #$systemd_dir = '/etc/systemd/system'
}
