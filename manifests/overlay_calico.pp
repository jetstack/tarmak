class puppernetes::overlay_calico {
  include ::puppernetes
  require ::vault_client

  $etcd_overlay_base_path = "${::puppernetes::etcd_ssl_dir}/${::puppernetes::etcd_overlay_ca_name}"
  vault_client::cert_service { 'etcd-overlay':
    base_path   => $etcd_overlay_base_path,
    common_name => "${::hostname}.${::puppernetes::cluster_name}.${::puppernetes::dns_root}",
    role        => "${::puppernetes::cluster_name}/pki/${::puppernetes::etcd_overlay_ca_name}/sign/client",
    ip_sans     => $::puppernetes::ipaddress,
  }

  class { 'calico':
    etcd_cluster     => $::puppernetes::_etcd_cluster,
    tls              => true,
    systemd_after    => ['etcd-overlay-cert.service'],
    systemd_requires => ['etcd-overlay-cert.service'],
  }

  Service['etcd-overlay-cert.service'] -> Service['calico-node.service']

  if $::puppernetes::role == 'master' {
    class { 'calico::policy_controller': }

    calico::ip_pool {$::puppernetes::kubernetes_pod_network:
      ip_pool      => $::puppernetes::kubernetes_pod_network_host,
      ip_mask      => $::puppernetes::kubernetes_pod_network_mask,
      ipip_enabled => 'true', #lint:ignore:quoted_booleans
    }
  }
}
