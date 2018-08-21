class site_module::docker_storage(
  String $vg_name = 'vg_docker',
  String $vg_initial_size = '60%FREE',
){
  include ::site_module

  $conf_file = '/etc/sysconfig/docker-storage-setup'

  $ebs_device = $::site_module::ebs_device

  file {$conf_file:
    ensure  => file,
    content => template('site_module/docker-storage-setup.erb'),
    before  => Service['docker.service'],
    require => Package['docker'],
  }
}
