# Class: kubernetes
class kubernetes (
  $version = $::kubernetes::params::version,
  $bin_dir = $::kubernetes::params::bin_dir,
  $download_dir = $::kubernetes::params::download_dir,
  $dest_dir = $::kubernetes::params::dest_dir,
  $config_dir = $::kubernetes::params::config_dir,
  $systemd_dir = $::kubernetes::params::systemd_dir,
  $run_dir = $::kubernetes::params::run_dir,
  $apply_dir = $::kubernetes::params::apply_dir,
  $uid = $::kubernetes::params::uid,
  $gid = $::kubernetes::params::gid,
  $user = $::kubernetes::params::user,
  $group = $::kubernetes::params::group,
  $master_url = $::kubernetes::params::master_url,
  $curl_path = $::kubernetes::params::curl_path,
  $ssl_dir = undef,
  $source = undef,
  Enum['aws', ''] $cloud_provider = '',
  $cluster_name = undef,
  $dns_root = undef,
  $cluster_dns = undef,
  $cluster_domain = 'cluster.local',
  $service_ip_range_network = '10.254.0.0',
  $service_ip_range_mask = '16',
  $leader_elect = true,
  $allow_privileged = true,
  $service_account_key_file = undef,
  $service_account_key_generate = false,
  Array[Enum['AlwaysAllow', 'ABAC', 'RBAC']] $authorization_mode = [],
) inherits ::kubernetes::params
{

  # detect authorization mode
  if $authorization_mode == [] {
    # enable RBAC after 1.6+
    if versioncmp($::kubernetes::version, '1.6.0') >= 0 {
      $_authorization_mode = ['ABAC','RBAC']
    } else {
      $_authorization_mode = ['ABAC']
    }
  } else {
    $_authorization_mode = $authorization_mode
  }

  $download_url = regsubst(
    $::kubernetes::params::download_url,
    '#VERSION#',
    $version,
    'G'
  )
  $_dest_dir = "${dest_dir}/kubernetes-${version}"

  if $ssl_dir == undef {
    $_ssl_dir = "${config_dir}/ssl"
  } else {
    $_ssl_dir = $ssl_dir
  }

  if $service_account_key_file == undef {
    $_service_account_key_file = "${_ssl_dir}/service-account-key.pem"
  } else {
    $_service_account_key_file = $service_account_key_file
  }

  if $cluster_dns == undef {
    $_sir_parts = $service_ip_range_network.split('\.')
    $_cluster_dns = "${_sir_parts[0]}.${_sir_parts[1]}.${_sir_parts[2]}.10"
  } else {
    $_cluster_dns = $cluster_dns
  }

  $path = defined('$::path') ? {
      default => '/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/opt/bin',
      true    => $::path
  }

  $nologin = $::osfamily ? {
    'RedHat' => '/sbin/nologin',
    'Debian' => '/usr/sbin/nologin',
    default  => '/usr/sbin/nologin',
  }

  group { $group:
    ensure => present,
    gid    => $gid,
  } ->
  user { $user:
    ensure => present,
    uid    => $uid,
    home   => $config_dir,
    shell  => $nologin,
  }

  file { $config_dir:
    ensure  => directory,
    owner   => $user,
    group   => $group,
    mode    => '0750',
    require => User[$user],
  } ->
  file { $_ssl_dir:
    ensure => directory,
    owner  => $user,
    group  => $group,
    mode   => '0750',
  }

  file {$::kubernetes::params::run_dir:
    ensure  => directory,
    owner   => $user,
    group   => $group,
    mode    => '0750',
    require => User[$user],
  }

  file {$::kubernetes::params::apply_dir:
    ensure  => directory,
    owner   => $user,
    group   => $group,
    mode    => '0750',
    require => User[$user],
  }

}
