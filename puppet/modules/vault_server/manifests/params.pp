# == Class vault_server::params
#
# This class is meant to be called from vault_server.
# It sets variables according to platform.
#
class vault_server::params {
  $app_name = 'vault'
  $version = '0.9.5'
  $unsealer_version = '0.3.0'
  $bin_dir = '/opt/bin'
  $dest_dir = '/opt'
  $config_dir = '/etc/vault'
  $lib_dir = '/var/lib/vault'
  $download_dir = '/tmp'
  $download_url = 'https://releases.hashicorp.com/vault/#VERSION#/vault_#VERSION#_linux_amd64.zip'
  $sha256sums_url = 'https://releases.hashicorp.com/vault/#VERSION#/vault_#VERSION#_SHA256SUMS'
  $signature_url = 'https://releases.hashicorp.com/vault/#VERSION#/vault_#VERSION#_SHA256SUMS.sig'
  $unsealer_download_url = 'https://github.com/jetstack/vault-unsealer/releases/download/#VERSION#/vault-unsealer_#VERSION#_linux_amd64'
  $unsealer_sha256 = 'c73783f7856190caa3140dcf1d3670c76da669a251f3ef9cfcb838b44cd9f4b3'
  $server_url = 'http://127.0.0.1:8200'
  $systemd_dir = '/etc/systemd/system'
}
