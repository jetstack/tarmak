class puppernetes::single_node(
){
  include '::puppernetes::etcd'
  ensure_resource('class', '::puppernetes::master',{
    disable_kubelet => true,
    disable_proxy   => true,
  })

  ensure_resource('class', '::puppernetes::worker',{})

  Class['puppernetes::etcd'] ->
  Class['puppernetes::master']
}
