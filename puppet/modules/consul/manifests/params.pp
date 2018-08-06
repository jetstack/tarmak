# == Class consul::params
#
# This class is meant to be called from consul.
# It sets variables according to platform.
#
class consul::params {
    $app_name = 'consul'
    $version = '1.2.1'
    $bin_dir = '/usr/local/bin'
    $consul_config_dir = '/etc/consul'
    $vault_config_dir = '/etc/vault'
    $download_dir = '/tmp'
    $systemd_dir = '/etc/systemd/system'
}
