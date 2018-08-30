# == Class vault_server::params
#
# This class is meant to be called from vault_server.
# It sets variables according to platform.
#
class vault_server::params {
  $app_name = 'vault'
  $version = '0.9.5'
  $unsealer_version = '0.1.4'
  $bin_dir = '/opt/bin'
  $dest_dir = '/opt'
  $config_dir = '/etc/vault'
  $lib_dir = '/var/lib/vault'
  $download_dir = '/tmp'
  $download_url = 'https://releases.hashicorp.com/vault/#VERSION#/vault_#VERSION#_linux_amd64.zip'
  $sha256sums_url = 'https://releases.hashicorp.com/vault/#VERSION#/vault_#VERSION#_SHA256SUMS'
  $signature_url = 'https://releases.hashicorp.com/vault/#VERSION#/vault_#VERSION#_SHA256SUMS.sig'
  $unsealer_download_url = 'https://github.com/jetstack/vault-unsealer/releases/download/#VERSION#/vault-unsealer_#VERSION#_linux_amd64'
  $unsealer_sha256 = '7a01a119429b93edecb712aa897f2b22ba0575b7db5f810d4a9a40d993dad1aa'
  $server_url = 'http://127.0.0.1:8200'
  $systemd_dir = '/etc/systemd/system'
}
