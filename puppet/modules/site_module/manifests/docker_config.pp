class site_module::docker_config {
  file { '/etc/sysconfig/docker':
    ensure  => file,
    content => template('site_module/docker.erb'),
  }
  file { '/etc/systemd/system/docker.service.d':
    ensure  => directory,
  } -> file { '/etc/systemd/system/docker.service.d/10-slice.conf':
    ensure  => directory,
    content => '[Service]\nSlice=podruntime.slice\n',
  }
}
