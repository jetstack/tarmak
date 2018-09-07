class vault_server (
  String $region = '',
  String $environment = '',
  String $vault_tls_cert_path = '',
  String $vault_tls_ca_path = '',
  String $vault_tls_key_path = '',
  Optional[String] $vault_unsealer_kms_key_id = undef,
  Optional[String] $vault_unsealer_ssm_key_prefix = undef,
  Optional[String] $consul_master_token = undef,
  Optional[String] $vault_unsealer_key_dir = $vault_server::params::config_dir,
  Enum['aws', ''] $cloud_provider = '',
  String $app_name = $vault_server::params::app_name,
  String $version = $vault_server::params::version,
  String $bin_dir = $vault_server::params::bin_dir,
  String $dest_dir = $vault_server::params::dest_dir,
  String $config_dir = $vault_server::params::config_dir,
  String $lib_dir = $vault_server::params::lib_dir,
  String $download_dir = $vault_server::params::download_dir,
  String $systemd_dir = $vault_server::params::systemd_dir,
) inherits ::vault_server::params {

  # paths
  $path = defined('$::path') ? {
    default => '/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/opt/bin',
    true    => $::path,
  }

  $_consul_master_token = $consul_master_token ? {
    undef   => $::consul_master_token,
    default => $consul_master_token,
  }

  $_dest_dir = "${dest_dir}/${app_name}-${version}"
  $bin_path = "${_dest_dir}/${app_name}"
  $unsealer_bin_path = "${bin_path}-unsealer"
  $link_path = "${dest_dir}/bin"

  file { $config_dir:
    ensure => 'directory',
    mode   => '0777',
  }

  file { $lib_dir:
    ensure => 'directory',
    mode   => '0777',
  }

  user { $app_name:
    ensure => 'present',
    system => true,
    home   => $lib_dir,
  }

  # install airworthy if necessary
  if !defined(Class['::airworthy']) {
    class {'::airworthy':}
  }

  Class['::airworthy']
  ~> class { '::vault_server::install': }
  ~> class { '::vault_server::service': }
}
