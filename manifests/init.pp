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
  $ssl_dir = undef,
  $source = undef,
  $cloud_provider = undef,
  $cluster_name = undef,
  $dns_root = undef,
  $cluster_dns = undef,
  $cluster_domain = 'cluster.local',
  $service_ip_range_network = '10.254.0.0',
  $service_ip_range_mask = '16',
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

  file {$::kubernetes::params::apply_dir:
    ensure  => directory,
    owner   => $user,
    group   => $group,
    mode    => '0750',
    require => User[$user],
  }

}
