class calico::node(
  $aws_filter_hack = $::calico::params::aws_filter_hack,
  $node_version = $::calico::params::calico_node_version,
  $etcd_cert_path = $::calico::params::etcd_cert_path
) inherits ::calico
{

  include ::calico

  $download_url = regsubst(
    $::calico::params::calico_node_download_url,
    '#VERSION#',
    $node_version,
    'G')

  calico::wget_file { 'calicoctl':
    url             => "${$download_url}/calicoctl",
    destination_dir => "${::calico::install_dir}/bin",
    require         => Class['calico'],
    before          => File["${::calico::install_dir}/bin/calicoctl"],
  }

  file { "${::calico::install_dir}/bin/calicoctl":
    ensure => file,
    mode   => '0755',
  }

  file { "${::calico::config_dir}/calico.env":
    ensure  => file,
    content => template('calico/calico.env.erb'),
    require => Class['calico'],
  }

  file { "${::calico::systemd_dir}/calico-node.service":
    ensure  => file,
    content => template('calico/calico-node.service.erb'),
  } ~>
  exec { "${module_name}-systemctl-daemon-reload":
    command     => '/usr/bin/systemctl daemon-reload',
    refreshonly => true,
  }

  service { 'calico-node':
    ensure    => running,
    enable    => true,
    require   => [ File["${::calico::config_dir}/calico.env"], File["${::calico::systemd_dir}/calico-node.service"] ],
    subscribe => File["${::calico::config_dir}/calico.env"],
  }

  if $aws_filter_hack {
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
}
