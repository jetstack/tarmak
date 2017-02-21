class puppernetes::single_node(
  $dns_root = 'jetstack.net',
  $cluster_name = 'cluster',
){
  ensure_resource('class', '::puppernetes',{
    cluster_name   => $cluster_name,
    dns_root       => $dns_root,
    etcd_cluster   => ["${::hostname}.${cluster_name}.${dns_root}"],
    etcd_instances => 1,
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
