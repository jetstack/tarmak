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
  String $consul_encrypt = '',
  String $fqdn = '',
  String $private_ip = '127.0.0.1',
  String $consul_master_token = '',
  String $region = '',
  String $instance_count = '',
  String $environment = '',
  String $backup_bucket_prefix = '',
  String $backup_schedule = '*-*-* 00:00:00',
  String $volume_id = '',
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
  Integer $uid = 871,
  Integer $gid = 871,
  String $user = 'consul',
  String $group = 'consul',
  Enum['aws', ''] $cloud_provider = '',
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
  $bin_path = "${_dest_dir}/${app_name}"
  $link_path = "${dest_dir}/bin"
  $config_path = "${config_dir}/consul.json"

  $exporter_dest_dir = "${dest_dir}/${app_name}_exporter-${exporter_version}"
  $exporter_bin_path = "${exporter_dest_dir}/${app_name}_exporter"

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
  ensure_resource('class', '::airworthy', {})

  Class['::airworthy']
  ~> class { '::consul::install': }
  ~> class { '::consul::config': }
  ~> class { '::consul::service': }
}
