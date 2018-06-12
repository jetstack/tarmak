class site_module::docker_config {
  file { '/etc/sysconfig/docker':
    ensure  => file,
    content => template('site_module/docker.erb'),
  }
  file { "/etc/systemd/system/${::site_module::docker::service_name}.service.d":
    ensure  => directory,
  } -> file { '/etc/systemd/system/docker.service.d/10-slice.conf':
    ensure  => file,
    content => "[Service]\nSlice=podruntime.slice\n",
    notify  => Service["${::site_module::docker::service_name}.service"],
  } -> file { '/etc/systemd/system/docker.service.d/20-cgroupfs.conf':
    ensure  => file,
    content => file('site_module/20-cgroupfs.conf'),
    notify  => Service["${::site_module::docker::service_name}.service"],
  }
}
