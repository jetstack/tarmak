# download and install hyperkube
class kubernetes::install (
  Integer $max_user_instances = $::kubernetes::max_user_instances,
  Integer $max_user_watches = $::kubernetes::max_user_watches,
){
  include kubernetes

  $hyperkube_path = "${::kubernetes::_dest_dir}/hyperkube"

  file { $::kubernetes::_dest_dir:
    ensure => directory,
    mode   => '0755',
  }
  -> exec {"kubernetes-${kubernetes::version}-download":
    command => "curl -sL -o  ${hyperkube_path} ${::kubernetes::download_url}",
    creates => $hyperkube_path,
    path    => ['/usr/bin/', '/bin'],
  }
  -> file {"${::kubernetes::_dest_dir}/hyperkube":
    ensure => file,
    mode   => '0755',
    owner  => 'root',
    group  => 'root',
  }

  exec {'sysctl-system':
    command     => 'sysctl --system',
    refreshonly => true,
    path        => ['/usr/bin/', '/bin', '/usr/sbin'],
  }

  file{"${::kubernetes::params::sysctl_dir}/fs.conf":
    ensure  => file,
    mode    => '0640',
    owner   => 'root',
    content => template('kubernetes/fs.conf.erb'),
    notify  => Exec['sysctl-system'],
  }
}
