class site_module::docker_config {
  file { '/etc/sysconfig/docker':
    ensure  => file,
    content => template('site_module/docker.erb'),
  }
}
