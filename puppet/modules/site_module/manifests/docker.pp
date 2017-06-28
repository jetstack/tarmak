class site_module::docker{

  package{'docker':
    ensure  => present,
  }
  -> class{'site_module::docker_config':}
  -> service{'docker.service':
    ensure => running,
    enable => true,
  }

  if defined(Class['kubernetes::kubelet']){
    Class['site_module::docker'] -> Class['kubernetes::kubelet']
  }

  if defined(Class['prometheus']){
    Class['site_module::docker'] -> Class['prometheus']
  }

}
