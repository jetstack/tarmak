class calico::node(
  $etcd_endpoints,
  $etcd_cert_file,
  $etcd_key_file,
  $etcd_ca_file,
  $aws_filter_hack = $::calico::params::aws_filter_hack,
  $node_version = $::calico::params::calico_node_version,
  $etcd_cert_path = $::calico::params::etcd_cert_path
) inherits ::calico::params
{
  $download_url = regsubst(
    $::calico::params::calico_node_download_url,
    '#VERSION#',
    $node_version,
    'G')

  file { "${::calico::params::cni_base_dir}/cni/net.d/10-calico.conf":
    ensure  => file,
    content => template('calico/10-calico.conf.erb'),
  }

  calico::wget_file { 'calicoctl':
    url             => "${$download_url}/calicoctl",
    destination_dir => "${::calico::params::install_dir}/bin",
    before          => File["${::calico::params::install_dir}/bin/calicoctl"],
  }

  file { "${::calico::params::install_dir}/bin/calicoctl":
    ensure => file,
    mode   => '0755',
  }

  file { "${::calico::params::config_dir}/calico.env":
    ensure  => file,
    content => template('calico/calico.env.erb'),
  }

  file { "${::calico::params::systemd_dir}/calico-node.service":
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
    require   => [ File["${::calico::params::config_dir}/calico.env"], File["${::calico::params::systemd_dir}/calico-node.service"] ],
    subscribe => File["${::calico::params::config_dir}/calico.env"],
  }

  if $aws_filter_hack {
    file { "${::calico::params::helper_dir}/calico_filter_hack.sh":
      ensure  => file,
      content => template('calico/calico_filter_hack.sh.erb'),
      mode    => '0750',
    } ->
    exec { 'Modify calico filter':
      command => "${::calico::params::helper_dir}/calico_filter_hack.sh set",
      unless  => "${::calico::params::helper_dir}/calico_filter_hack.sh test",
      require => Service['calico-node'],
    }
  }
}
