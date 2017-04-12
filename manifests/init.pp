# calico init.pp
class calico(
  Array[String] $etcd_cluster = $::calico::params::etcd_cluster,
  Integer[1,65535] $etcd_overlay_port = $::calico::params::etcd_overlay_port,
  Enum['etcd', 'kubernetes'] $backend = 'etcd',
  String $etcd_ca_file = '',
  String $etcd_cert_file = '',
  String $etcd_key_file = '',
  String $cloud_provider = $::calico::params::cloud_provider,
  String $namespace = 'kube-system',
) inherits ::calico::params
{
  $path = defined('$::path') ? {
      default => '/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/opt/bin',
      true    => $::path
  }

  if $backend == 'etcd' {
    if $etcd_ca_file != '' and $etcd_cert_file != '' and $etcd_key_file != '' {
      $etcd_proto = 'https'

      if dirname($etcd_ca_file) != dirname($etcd_key_file) or dirname($etcd_key_file) != dirname($etcd_cert_file) {
        fail('etcd_key_file, etcd_cert_file and etcd_ca_file must be stored in the same directory')
      }
      $etcd_tls_dir = dirname($etcd_ca_file)
    } else {
      $etcd_proto = 'http'
    }
    $etcd_endpoints = $etcd_cluster.map |$node| { "${etcd_proto}://${node}:${etcd_overlay_port}" }.join(',')
  } elsif $backend == 'kubernetes' {
    fail('Backend storage kubernetes is not yet supported')
  }

  if $cloud_provider == 'aws' {
    include ::calico::disable_source_destination_check
  }

  # make sure old stuff is disabled
  $node_service_name = 'calico-node.service'
  exec {"systemctl stop ${node_service_name}":
    onlyif => "test -f ${calico::systemd_dir}/${node_service_name}",
    path   => $path,
  } ->
  file{"${calico::systemd_dir}/${node_service_name}":
    ensure =>  absent,
  }
}
