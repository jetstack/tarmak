class puppernetes::single_node(
  String $dns_root = $puppernetes::params::dns_root,
  String $cluster_name = $puppernetes::params::cluster_name,
  String $etcd_advertise_client_network = $puppernetes::params::etcd_advertise_client_network,
  String $kubernetes_api_url = nil,
  String $kubernetes_version = $puppernetes::params::kubernetes_version,
) inherits puppernetes::params{
  ensure_resource('class', '::puppernetes',{
    dns_root                      => $dns_root,
    cluster_name                  => $cluster_name,
    etcd_advertise_client_network => $etcd_advertise_client_network,
    kubernetes_api_url            => $kubernetes_api_url,
    kubernetes_version            => $kubernetes_version,
    etcd_cluster                  => ["${::hostname}.${cluster_name}.${dns_root}"],
    etcd_instances                => 1,
  })

  include '::puppernetes::etcd'

  ensure_resource('class', '::puppernetes::master',{
    disable_kubelet => true,
    disable_proxy   => true,
  })

  ensure_resource('class', '::puppernetes::worker',{})

  Class['puppernetes::etcd'] ->
  Class['puppernetes::master']
}
