class calico::node
{

  include ::calico

  include k8s

  $version = $::calico::params::calico_node_version

  $download_url = regsubst(
    $::calico::params::calico_node_download_url,
    '#VERSION#',
    $version,
    'G')

  wget::fetch { "calicoctl-v${version}":
    source      => "${$download_url}/calicoctl",
    destination => "${::calico::install_dir}/bin/",
    require     => Class['calico'],
    before      => File["${::calico::install_dir}/bin/calicoctl"],
  }

  file { "${::calico::install_dir}/bin/calicoctl":
    ensure => file,
    mode   => '0755',
  }

  file { "${::calico::conf_dir}/calico.env":
    ensure  => file,
    content => template('calico/calico.env.erb'),
    require => Class['calico'],
  }

  file { "${::calico::systemd_dir}/calico-node.service":
    ensure  => file,
    content => template('calico/calico-node.service.erb'),
  } ~>
  Exec["${module_name}-systemctl-daemon-reload"]

  service { 'calico-node':
    ensure    => running,
    enable    => true,
    require   => [ Class['k8s'], File["${::calico::conf_dir}/calico.env"], File["${::calico::systemd_dir}/calico-node.service"] ],
    subscribe => File["${::calico::conf_dir}/calico.env"],
  }

  file { "${::calico::helper_dir}/calico_filter_hack.sh":
    ensure  => file,
    content => template('calico/calico_filter_hack.sh.erb'),
    mode    => '0750',
  } ->
  exec { 'Modify calico filter':
    command => "${::calico::helper_dir}/calico_filter_hack.sh set",
    unless  => "${::calico::helper_dir}/calico_filter_hack.sh test",
    require => Service['calico-node'],
  }
}
