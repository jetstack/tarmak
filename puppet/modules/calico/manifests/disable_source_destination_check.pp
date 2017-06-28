# This class disable the source/destination check on AWS instances
class calico::disable_source_destination_check{
  include calico

  $service_name = 'disable-source-destination-check.service'
  $bin_path = "${::calico::bin_dir}/disable-source-destination-check.sh"

  file {$bin_path:
    ensure  => file,
    content => template('calico/disable-source-destination-check.sh.erb'),
    mode    => '0755',
    notify  => Service[$service_name]
  }

  file {"${::calico::systemd_dir}/${service_name}":
    ensure  => file,
    content => template('calico/disable-source-destination-check.service.erb'),
    notify  => Service[$service_name]
  } ~> exec { "${service_name}-daemon-reload":
    command     => 'systemctl daemon-reload',
    path        => $calico::path,
    refreshonly => true,
  } -> service{$service_name:
    ensure => running,
    enable => true,
  }

}
