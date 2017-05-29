class puppernetes::overlay_calico {
  include ::puppernetes
  require ::vault_client

  ensure_resource('file', $::puppernetes::etcd_home, {'ensure'    => 'directory' })
  ensure_resource('file', $::puppernetes::etcd_ssl_dir, {'ensure' => 'directory' })

  $etcd_overlay_base_path = "${::puppernetes::etcd_ssl_dir}/${::puppernetes::etcd_overlay_ca_name}"
  vault_client::cert_service { 'etcd-overlay':
    base_path   => $etcd_overlay_base_path,
    common_name =>  'etcd-client',
    role        => "${::puppernetes::cluster_name}/pki/${::puppernetes::etcd_overlay_ca_name}/sign/client",
    ip_sans     => [$::puppernetes::ipaddress],
    alt_names   => ["${::hostname}.${::puppernetes::cluster_name}.${::puppernetes::dns_root}"],
    exec_post   => [
      "-${::puppernetes::systemctl_path} --no-block try-restart calico-node.service",
      "-/bin/bash -c 'docker ps -q --filter=label=io.kubernetes.container.name=calico-policy-controller | xargs docker kill'",
    ],
  }

  class { 'calico':
    etcd_cluster   => $::puppernetes::_etcd_cluster,
    etcd_ca_file   => "${etcd_overlay_base_path}-ca.pem",
    etcd_cert_file => "${etcd_overlay_base_path}.pem",
    etcd_key_file  => "${etcd_overlay_base_path}-key.pem",
  }

  File[$::puppernetes::etcd_home] -> File[$::puppernetes::etcd_ssl_dir] -> Service['etcd-overlay-cert.service']

  if $::puppernetes::role == 'master' {
    class { 'calico::config': }
    Class['calico::config'] -> class { 'calico::policy_controller': }
    Class['calico::config'] -> class { 'calico::node': }
  }
}
