class puppernetes::etcd(
){
  include ::puppernetes
  include ::vault_client
  Class['vault_client'] -> Class['puppernetes::etcd']

  $initial_cluster = range(0, $::puppernetes::etcd_instances-1).map |$i| { #lint:ignore:variable_contains_dash
    "etcd-${i}.${::puppernetes::cluster_name}.${::puppernetes::dns_root}"
  }

  $nodename                 = "${::hostname}.${::puppernetes::cluster_name}.${::puppernetes::dns_root}"

  class{'etcd':
    user  => $::puppernetes::etcd_user,
    group => $::puppernetes::etcd_group
  }

  etcd::instance{'k8s-main':
    version                  => $::puppernetes::etcd_k8s_main_version,
    nodename                 => $nodename,
    members                  => $::puppernetes::etcd_instances,
    initial_cluster          => $initial_cluster,
    advertise_client_network => $::puppernetes::etcd_advertise_client_network,
    client_port              => $::puppernetes::etcd_k8s_main_client_port,
    peer_port                => $::puppernetes::etcd_k8s_main_peer_port,
    tls                      => true,
    tls_cert_path            => "${::puppernetes::etcd_ssl_dir}/etcd-${::puppernetes::etcd_k8s_main_ca_name}.pem",
    tls_key_path             => "${::puppernetes::etcd_ssl_dir}/etcd-${::puppernetes::etcd_k8s_main_ca_name}-key.pem",
    tls_ca_path              => "${::puppernetes::etcd_ssl_dir}/etcd-${::puppernetes::etcd_k8s_main_ca_name}-ca.pem",
  }
  etcd::instance{'k8s-events':
    version                  => $::puppernetes::etcd_k8s_events_version,
    nodename                 => $nodename,
    members                  => $::puppernetes::etcd_instances,
    initial_cluster          => $initial_cluster,
    advertise_client_network => $::puppernetes::etcd_advertise_client_network,
    client_port              => $::puppernetes::etcd_k8s_events_client_port,
    peer_port                => $::puppernetes::etcd_k8s_events_peer_port,
    tls                      => true,
    tls_cert_path            => "${::puppernetes::etcd_ssl_dir}/etcd-${::puppernetes::etcd_k8s_events_ca_name}.pem",
    tls_key_path             => "${::puppernetes::etcd_ssl_dir}/etcd-${::puppernetes::etcd_k8s_events_ca_name}-key.pem",
    tls_ca_path              => "${::puppernetes::etcd_ssl_dir}/etcd-${::puppernetes::etcd_k8s_events_ca_name}-ca.pem",
  }
  etcd::instance{'overlay':
    version                  => $::puppernetes::etcd_overlay_version,
    nodename                 => $nodename,
    members                  => $::puppernetes::etcd_instances,
    initial_cluster          => $initial_cluster,
    advertise_client_network => $::puppernetes::etcd_advertise_client_network,
    client_port              => $::puppernetes::etcd_overlay_client_port,
    peer_port                => $::puppernetes::etcd_overlay_peer_port,
    tls                      => true,
    tls_cert_path            => "${::puppernetes::etcd_ssl_dir}/etcd-${::puppernetes::etcd_overlay_ca_name}.pem",
    tls_key_path             => "${::puppernetes::etcd_ssl_dir}/etcd-${::puppernetes::etcd_overlay_ca_name}-key.pem",
    tls_ca_path              => "${::puppernetes::etcd_ssl_dir}/etcd-${::puppernetes::etcd_overlay_ca_name}-ca.pem",
  }
}
