# Class: kubernetes
class kubernetes (
  $version = $::kubernetes::params::version,
  $bin_dir = $::kubernetes::params::bin_dir,
  $download_dir = $::kubernetes::params::download_dir,
  $dest_dir = $::kubernetes::params::dest_dir,
  $config_dir = $::kubernetes::params::config_dir,
  $systemd_dir = $::kubernetes::params::systemd_dir,
  $run_dir = $::kubernetes::params::run_dir,
  $uid = $::kubernetes::params::uid,
  $gid = $::kubernetes::params::gid,
  $user = $::kubernetes::params::user,
  $group = $::kubernetes::params::group,
  $master_url = $::kubernetes::params::master_url,
  $ssl_dir = undef,
  $source = undef,
  $cloud_provider = undef,
  $cluster_name = undef,
  $dns_root = undef,
  $cluster_dns = 'cluster.local',
  $cluster_ip = '10.254.0.0',
  $cluster_ip_mask = 16,
  $leader_elect = true,
  $allow_privileged = true,
  $service_account_key_file = undef,
) inherits ::kubernetes::params
{
  $download_url = regsubst(
    $::kubernetes::params::download_url,
    '#VERSION#',
    $version,
    'G'
  )
  $real_dest_dir = "${dest_dir}/kubernetes-${version}"

  if $ssl_dir == undef {
    $real_ssl_dir = "${config_dir}/ssl"
  } else {
    $real_ssl_dir = $ssl_dir
  }

  $path = defined('$::path') ? {
      default => '/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/opt/bin',
      true    => $::path
  }

  group { $group:
    ensure => present,
    gid    => $gid,
  } ->
  user { $user:
    ensure => present,
    uid    => $uid,
    shell  => '/sbin/nologin',
    home   => $config_dir,
  }

  file { $config_dir:
    ensure  => directory,
    owner   => $user,
    group   => $group,
    mode    => '0750',
    require => User[$user],
  } ->
  file { $real_ssl_dir:
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

}
