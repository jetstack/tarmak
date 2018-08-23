class site_module::docker{

  $service_name = 'docker'
  $path = defined('$::path') ? {
      default => '/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/opt/bin',
      true    => $::path
  }

  package{'docker':
    ensure  => present,
  }
  -> class{'site_module::docker_config':}
  ~> exec { "${service_name}-daemon-reload":
    command     => 'systemctl daemon-reload',
    path        => $path,
    refreshonly => true,
  }
  -> service{"${service_name}.service":
    ensure     => running,
    enable     => true,
    hasstatus  => true,
    hasrestart => true,
  }

  if defined(Class['kubernetes::kubelet']){
    Class['site_module::docker'] -> Class['kubernetes::kubelet']
  }

  if defined(Class['prometheus']){
    Class['site_module::docker'] -> Class['prometheus']
  }

}
