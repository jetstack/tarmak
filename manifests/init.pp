# calico init.pp
class calico(
  $etcd_cluster = $::calico::params::etcd_cluster,
  $etcd_overlay_port = $::calico::params::etcd_overlay_port,
  $tls = $::calico::params::tls,
  $aws = $::calico::params::aws,
  $aws_filter_hack = $::calico::params::aws_filter_hack,
  $systemd_wants = $::calico::params::systemd_wants,
  $systemd_requires = $::calico::params::systemd_requires,
  $systemd_after = $::calico::params::systemd_after,
  $systemd_before = $::calico::params::systemd_before,
) inherits ::calico::params
{
  if $tls {
    $proto = 'https'
  } else {
    $proto = 'http'
  }

  $etcd_cert_file = "${::calico::params::etcd_cert_path}/${::calico::params::etcd_cert_base_name}.pem"
  $etcd_key_file  = "${::calico::params::etcd_cert_path}/${::calico::params::etcd_cert_base_name}-key.pem"
  $etcd_ca_file   = "${::calico::params::etcd_cert_path}/${::calico::params::etcd_cert_base_name}-ca.pem"

  $etcd_endpoints = $etcd_cluster.map |$node| { "${proto}://${node}:${etcd_overlay_port}" }.join(',')

  file { [
    "${::calico::cni_base_dir}/cni",
    "${::calico::cni_base_dir}/cni/net.d",
    $::calico::config_dir, $::calico::install_dir,
    "${::calico::install_dir}/bin"
  ]:
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

  file { "${::calico::helper_dir}/calico_helper.sh":
    ensure  => file,
    content => template('calico/calico_helper.sh.erb'),
    mode    => '0755',
  }

  class {'::calico::bin_install':} ->
  class {'::calico::lo_install':} ->
  class {'::calico::node':
    etcd_endpoints   => $etcd_endpoints,
    etcd_cert_file   => $etcd_cert_file,
    etcd_key_file    => $etcd_key_file,
    etcd_ca_file     => $etcd_ca_file,
    aws_filter_hack  => $aws_filter_hack,
    tls              => $tls,
    systemd_wants    => $systemd_wants,
    systemd_requires => ['docker.service'] + $systemd_requires,
    systemd_after    => ['docker.service'] + $systemd_after,
    systemd_before   => $systemd_before,
  }
}
