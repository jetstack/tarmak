# calico init.pp

class calico(
  $etcd_cluster,
  $etcd_overlay_port = $::calico::params::etcd_overlay_port,
  $tls = $::calico::params::tls,
  $aws = $::calico::params::aws,
) inherits ::calico::params
{
  if $tls {
    $proto = 'https'
  } else {
    $proto = 'http'
  }

  $etcd_cert_file = "${::calico::etcd_cert_path}/${::calico::etcd_cert_base_name}.pem"
  $etcd_key_file = "${::calico::etcd_cert_path}/${::calico::etcd_cert_base_name}-key.pem"
  $etcd_ca_file = "${::calico::etcd_cert_path}/${::calico::etcd_cert_base_name}-ca.pem"

  $etcd_endpoints = $etcd_cluster.map |$node| { "${proto}://${node}:${etcd_overlay_port}" }.join(',')

  file { ["${::calico::cni_base_dir}/cni", "${::calico::cni_base_dir}/cni/net.d", $::calico::config_dir, $::calico::install_dir, "${::calico::install_dir}/bin"]:
    ensure => directory,
  }

  if $aws {
    file { "${::calico::helper_dir}/sourcedestcheck.sh":
      ensure  => file,
      content => template('calico/sourcedestcheck.sh.erb'),
      mode    => '0755',
    }

    exec { 'Disable source dest check':
      command => "${::calico::helper_dir}/sourcedestcheck.sh set",
      unless  => "${::calico::helper_dir}/sourcedestcheck.sh test",
      require => File["${::calico::helper_dir}/sourcedestcheck.sh"],
    }
  }

  $path = defined('$::path') ? {
    default => '/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/opt/bin',
    true    => $::path
  }

  file { "${::calico::helper_dir}/calico_helper.sh":
    ensure  => file,
    content => template('calico/calico_helper.sh.erb'),
    mode    => '0755',
  }

  exec { "${module_name}-systemctl-daemon-reload":
    command     => 'systemctl daemon-reload',
    refreshonly => true,
    path        => $path,
  }
}
