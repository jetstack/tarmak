# Install/configure an etcd node.
#
# @param data_dir The directory to store etcd data
# @param config_dir The directory to store etcd config
# @param user The username to run etcd
# @param uid The user ID to run etcd
# @param group The group to run etcd
# @param gid The etcd group ID
class etcd(
  $data_dir = $::etcd::params::data_dir,
  $config_dir = $::etcd::params::config_dir,
  $uid = $::etcd::params::uid,
  $gid = $::etcd::params::gid,
  $user = $::etcd::params::user,
  $group = $::etcd::params::group,
  String $backup_bucket_prefix = '',
) inherits ::etcd::params {

  $nologin = $::osfamily ? {
    'RedHat' => '/sbin/nologin',
    'Debian' => '/usr/sbin/nologin',
    default  => '/usr/sbin/nologin',
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

  file { $data_dir:
    ensure  => directory,
    owner   => $user,
    group   => $group,
    mode    => '0750',
    require => User[$user],
  }

  file { $config_dir:
    ensure  => directory,
    owner   => $user,
    group   => $group,
    mode    => '0750',
    require => User[$user],
  }
}
