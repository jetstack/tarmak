# TODO Full description of class etcd here.
class etcd ($use_vault = true)
{

  if ($use_vault == true) {
    require ::vault_client
  }

  else {

    user { 'etcd':
      ensure => present,
      uid    => 873,
      shell  => '/sbin/nologin',
      home   => '/var/lib/etcd',
    }

    file { '/etc/etcd':
      ensure => directory,
      owner  => 'etcd',
      group  => 'etcd',
    }
  }

  file { '/var/lib/etcd':
    ensure  => directory,
    owner   => 'etcd',
    group   => 'etcd',
    require => User['etcd'],
  }
}
