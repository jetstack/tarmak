class tarmak::overlay_calico {
  include ::tarmak
  require ::vault_client

  ensure_resource('file', $::tarmak::etcd_home, {'ensure'    => 'directory' })
  ensure_resource('file', $::tarmak::etcd_ssl_dir, {'ensure' => 'directory' })

  $etcd_overlay_base_path = "${::tarmak::etcd_ssl_dir}/${::tarmak::etcd_overlay_ca_name}"
  vault_client::cert_service { 'etcd-overlay':
    base_path   => $etcd_overlay_base_path,
    common_name =>  'etcd-client',
    role        => "${::tarmak::cluster_name}/pki/${::tarmak::etcd_overlay_ca_name}/sign/client",
    ip_sans     => [$::tarmak::ipaddress],
    alt_names   => ["${::hostname}.${::tarmak::cluster_name}.${::tarmak::dns_root}"],
    uid         => $::tarmak::etcd_uid,
    exec_post   => [
      "-${::tarmak::systemctl_path} --no-block try-restart calico-node.service",
      "-/bin/bash -c 'docker ps -q --filter=label=io.kubernetes.container.name=calico-policy-controller | xargs docker kill'",
    ],
  }

  class { 'calico':
    etcd_cluster   => $::tarmak::_etcd_cluster,
    etcd_ca_file   => "${etcd_overlay_base_path}-ca.pem",
    etcd_cert_file => "${etcd_overlay_base_path}.pem",
    etcd_key_file  => "${etcd_overlay_base_path}-key.pem",
    pod_network    => $::tarmak::kubernetes_pod_network,
  }

  File[$::tarmak::etcd_home] -> File[$::tarmak::etcd_ssl_dir] -> Service['etcd-overlay-cert.service']

  if $::tarmak::role == 'master' {
    class { 'calico::config': }
    Class['calico::config'] -> class { 'calico::policy_controller': }
    Class['calico::config'] -> class { 'calico::node': }
  }
}
