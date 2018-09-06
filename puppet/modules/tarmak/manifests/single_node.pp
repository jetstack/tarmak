class tarmak::single_node(
  String $dns_root = $tarmak::params::dns_root,
  String $cluster_name = $tarmak::params::cluster_name,
  String $etcd_advertise_client_network = $tarmak::params::etcd_advertise_client_network,
  String $kubernetes_api_url = nil,
  String $kubernetes_version = $tarmak::params::kubernetes_version,
  Array[Enum['AlwaysAllow', 'ABAC', 'RBAC']] $kubernetes_authorization_mode = [],
) inherits tarmak::params{

  # install airworthy if necessary
  if !defined(Class['::tarmak']) {
    class {'::tarmak':
      dns_root                      => $dns_root,
      cluster_name                  => $cluster_name,
      etcd_advertise_client_network => $etcd_advertise_client_network,
      kubernetes_api_url            => $kubernetes_api_url,
      kubernetes_version            => $kubernetes_version,
      kubernetes_authorization_mode => $kubernetes_authorization_mode,
      etcd_cluster                  => ["${::hostname}.${cluster_name}.${dns_root}"],
      etcd_instances                => 1,
    }
  }

  include '::tarmak::etcd'

  if !defined(Class['::tarmak::master']) {
    class {'::tarmak::master':
      disable_kubelet => true,
      disable_proxy   => true,
    }
  }

  if !defined(Class['::tarmak::worker']) {
    class {'::tarmak::worker':}
  }

  Class['tarmak::etcd'] -> Class['tarmak::master']
}
