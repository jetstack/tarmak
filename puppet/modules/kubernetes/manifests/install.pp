# download and install hyperkube
class kubernetes::install {
  include kubernetes

  $hyperkube_path = "${::kubernetes::_dest_dir}/hyperkube"

  file { $::kubernetes::_dest_dir:
    ensure => directory,
    mode   => '0755',
  }
  -> exec {"kubernetes-${kubernetes::version}-download":
    command => "curl -sL -o  ${hyperkube_path} ${::kubernetes::real_download_url}",
    creates => $hyperkube_path,
    path    => ['/usr/bin/', '/bin'],
  }
  -> file {"${::kubernetes::_dest_dir}/hyperkube":
    ensure => file,
    mode   => '0755',
    owner  => 'root',
    group  => 'root',
  }
}
