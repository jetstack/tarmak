# Install/configure an consul node.
#
# @param data_dir The directory to store consul data
# @param config_dir The directory to store consul config
# @param user The username to run consul
# @param uid The user ID to run consul
# @param group The group to run consul
# @param gid The consul group ID
# @param version Version of consul to deploy
# @param cloud_provider Select cloud provider for consul discovery
# @param exporter_enabled Enable/disable prometheus exporter
# @param exporter_version Version of prometheus exporter
# @param backup_enabled Enable/disable backup
# @param backup_version Version of backup
# @param advertise_network Specify network used for consul


class consul(
  Optional[String] $consul_master_token = undef,
  Optional[String] $consul_encrypt = undef,
  Optional[String] $consul_bootstrap_expect = undef,
  Enum['aws', ''] $cloud_provider = '',
  String $fqdn = '',
  String $private_ip = '127.0.0.1',
  String $region = '',
  String $environment = '',
  String $backup_bucket_prefix = '',
  String $backup_schedule = '*-*-* 00:00:00',
  String $app_name = $consul::params::app_name,
  String $version = $consul::params::version,
  String $config_dir = $consul::params::config_dir,
  String $download_dir = $consul::params::download_dir,
  String $systemd_dir = $consul::params::systemd_dir,
  String $exporter_version = $consul::params::exporter_version,
  String $dest_dir = $consul::params::dest_dir,
  String $data_dir = $consul::params::data_dir,
  String $download_url = $consul::params::download_url,
  String $sha256sums_url = $consul::params::sha256sums_url,
  String $exporter_download_url = $consul::params::exporter_download_url,
  String $exporter_signature_url = $consul::params::exporter_signature_url,
  String $backinator_version = $consul::params::backinator_version,
  String $backinator_download_url = $consul::params::backinator_download_url,
  String $backinator_sha256 = $consul::params::backinator_sha256,
  Integer $uid = 871,
  Integer $gid = 871,
  String $user = 'consul',
  String $group = 'consul',
  Boolean $exporter_enabled = true,
  Boolean $backup_enabled = true,
  String $backup_version = 'xx',
  String $acl_default_policy = 'deny',
  String $acl_down_policy = 'deny',
  Boolean $server = true,
  String $client_addr = '127.0.0.1',
  String $bind_addr = '0.0.0.0',
  String $log_level = 'INFO',
  String $datacenter = 'dc1',
  Optional[String] $advertise_network = undef,
  Optional[Array[String]] $retry_join = [''],
  Optional[String] $ca_file = undef,
  Optional[String] $cert_file = undef,
  Optional[String] $key_file = undef,
  Array[String] $systemd_wants = [],
  Array[String] $systemd_requires = [],
  Array[String] $systemd_after = [],
  Array[String] $systemd_before = [],
) inherits ::consul::params {

  $path = defined('$::path') ? {
    default => '/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/opt/bin',
    true    => $::path,
  }

  Exec { path => $path }

  $_dest_dir = "${dest_dir}/${app_name}-${version}"
  $_backinator_dest_dir = "${dest_dir}/${app_name}-backinator-${backinator_version}"
  $bin_path = "${_dest_dir}/${app_name}"
  $link_path = "${dest_dir}/bin"
  $config_path = "${config_dir}/consul.json"

  $exporter_dest_dir = "${dest_dir}/${app_name}-exporter-${exporter_version}"
  $exporter_bin_path = "${exporter_dest_dir}/${app_name}_exporter"

  $_consul_master_token = $consul_master_token ? {
    undef   => $::consul_master_token,
    default => $consul_master_token,
  }

  $_consul_bootstrap_expect = $consul_bootstrap_expect ? {
    undef => defined('$::consul_bootstrap_expect') ? {
      true    =>  $::consul_bootstrap_expect.scanf('%i')[0],
      default =>  1,
    },
    default => $consul_bootstrap_expect,
  }

  $_consul_encrypt = $consul_encrypt ? {
    undef   => $::consul_encrypt,
    default => $consul_encrypt,
  }

  $nologin = $::osfamily ? {
    'RedHat' => '/sbin/nologin',
    'Debian' => '/usr/sbin/nologin',
    default  => '/usr/sbin/nologin',
  }

  file { $consul::data_dir:
    ensure => directory,
    owner  => 'root',
    group  => $group,
    mode   => '0750',
  }

  group { $group:
    ensure => present,
    gid    => $gid,
  }
  -> user { $user:
    ensure => present,
    uid    => $uid,
    shell  => $nologin,
    home   => $data_dir,
  }

  # install airworthy if necessary
  if !defined(Class['::airworthy']) {
    class {'::airworthy':}
  }

  Class['::airworthy']
  ~> class { '::consul::install': }
  ~> class { '::consul::config': }
  ~> class { '::consul::service': }
}
