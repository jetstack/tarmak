class site_module::docker_storage(
  String $ebs_device = undef,
  String $vg_name = 'vg_docker',
  String $vg_initial_size = '60%FREE',
){

  $conf_file = '/etc/sysconfig/docker-storage-setup'

  file {$conf_file:
    ensure  => file,
    content => template('site_module/docker-storage-setup.erb'),
    before  => Service['docker.service'],
    require => Package['docker'],
  }
}
