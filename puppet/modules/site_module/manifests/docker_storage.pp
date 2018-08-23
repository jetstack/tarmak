class site_module::docker_storage(
  String $vg_name = 'vg_docker',
  String $vg_initial_size = '100%FREE',
){

  include ::aws_ebs

  $conf_file = '/etc/sysconfig/docker-storage-setup'

  $disks = aws_ebs::disks()

  case $disks.length {
    0: {$ebs_device = undef}
    1: {$ebs_device = $disks[0]}
    default: {$ebs_device = $disks[1]}
  }

  file {$conf_file:
    ensure  => file,
    content => template('site_module/docker-storage-setup.erb'),
    before  => Service['docker.service'],
    require => Package['docker'],
  }
}
