# == Class consul::params
#
# This class is meant to be called from consul.
# It sets variables according to platform.
#
class consul::params {
  $app_name = 'consul'
  $version = '1.2.1'
  $download_dir = '/tmp'
  $systemd_dir = '/etc/systemd/system'
  $data_dir = '/var/lib/consul'
  $config_dir = '/etc/consul'
  $dest_dir = '/opt'
  $exporter_version = '0.3.0'
  $download_url = 'https://releases.hashicorp.com/consul/#VERSION#/consul_#VERSION#_linux_amd64.zip'
  $sha256sums_url = 'https://releases.hashicorp.com/consul/#VERSION#/consul_#VERSION#_SHA256SUMS'
  $signature_url = 'https://releases.hashicorp.com/consul/#VERSION#/consul_#VERSION#_SHA256SUMS.sig'
  $exporter_download_url = 'https://github.com/prometheus/consul_exporter/releases/download/v#VERSION#/consul_exporter-#VERSION#.linux-amd64.tar.gz'
  $exporter_signature_url = 'https://releases.tarmak.io/signatures/consul_exporter/#VERSION#/consul_exporter-#VERSION#.linux-amd64.tar.gz.asc'
}
