class puppernetes::etcd(
){
  include ::puppernetes
  include ::vault_client
  Class['vault_client'] -> Class['puppernetes::etcd']

  $initial_cluster          = split(inline_template(@(END)),',')
<% array=[] -%>
<% @etcd_count.to_i.times do |i| -%>
<%  array<<"etcd-#{i}.#{@cluster_name}.#{@dns_root}" -%>
<% end -%>
<%= array.join(',') %>
|- END

  $nodename                 = "${::hostname}.${::puppernetes::cluster_name}.${::puppernetes::dns_root}"

  etcd::instance{'k8s-main':
    version                  => $::puppernetes::etcd_k8s_main_version,
    nodename                 => $nodename,
    members                  => $etcd_count,
    initial_cluster          => $initial_cluster,
    advertise_client_network => $advertise_client_network,
    client_port              => $::puppernetes::etcd_k8s_main_client_port,
    peer_port                => $::puppernetes::etcd_k8s_main_peer_port,
    tls                      => true,
    tls_cert_path            => "${::puppernetes::etcd_ssl_dir}/etcd-${::puppernetes::etcd_k8s_main_ca_name}.pem",
    tls_key_path             => "${::puppernetes::etcd_ssl_dir}/etcd-${::puppernetes::etcd_k8s_main_ca_name}-key.pem",
    tls_ca_path              => "${::puppernetes::etcd_ssl_dir}/etcd-${::puppernetes::etcd_k8s_main_ca_name}-ca.pem",
  }
  etcd::instance{'k8s-events':
    version                  => '3.0.15',
    nodename                 => $nodename,
    members                  => $etcd_count,
    initial_cluster          => $initial_cluster,
    advertise_client_network => $advertise_client_network,
    client_port              => $events_etcd_client_port,
    peer_port                => $events_etcd_peer_port,
    tls                      => true,
    tls_cert_path            => '/etc/etcd/ssl/etcd-events.pem',
    tls_key_path             => '/etc/etcd/ssl/etcd-events-key.pem',
    tls_ca_path              => '/etc/etcd/ssl/etcd-events-ca.pem',
  }
  etcd::instance{'k8s-overlay':
    version                  => '2.3.7',
    nodename                 => $nodename,
    members                  => $etcd_count,
    initial_cluster          => $initial_cluster,
    advertise_client_network => $advertise_client_network,
    client_port              => $calico_etcd_client_port,
    peer_port                => $calico_etcd_peer_port,
    tls                      => true,
    tls_cert_path            => '/etc/etcd/ssl/etcd-overlay.pem',
    tls_key_path             => '/etc/etcd/ssl/etcd-overlay-key.pem',
    tls_ca_path              => '/etc/etcd/ssl/etcd-overlay-ca.pem',
  }
}
