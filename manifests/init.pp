# calico init.pp

class calico(
  $helper_dir = $::calico::params::helper_dir,
  $config_dir = $::calico::params::config_dir,
  $secure_config_dir = $::calico::params::secure_config_dir,
  $cni_base_dir = $::calico::params::cni_base_dir,
  $install_dir = $::calico::params::install_dir,
  $tls = $::calico::params::tls,
  $etcd_cluster = []
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

  $etcd_endpoints = $etcd_cluster.map |$node| { "${proto}://${node}:${::calico::etcd_overlay_port}" }.join(',')


  file { [$cni_base_dir, "${cni_base_dir}/cni", "${cni_base_dir}/cni/net.d", $config_dir, $install_dir, "${install_dir}/bin"]:
    ensure => directory,
  }

  file { "${helper_dir}/sourcedestcheck.sh":
    ensure  => file,
    content => template('calico/sourcedestcheck.sh.erb'),
    mode    => '0755',
  }

  exec { 'Disable source dest check':
    command => "${helper_dir}/sourcedestcheck.sh set",
    unless  => "${helper_dir}/sourcedestcheck.sh test",
    require => File["${helper_dir}/sourcedestcheck.sh"],
  }

  $path = defined('$::path') ? {
    default => '/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/opt/bin',
    true    => $::path
  }

  exec { "${module_name}-systemctl-daemon-reload":
    command     => 'systemctl daemon-reload',
    refreshonly => true,
    path        => $path,
  }

  include ::calico::bin_install
  include ::calico::config
  include ::calico::lo_install
  include ::calico::node
  include ::calico::policy_controller
}
