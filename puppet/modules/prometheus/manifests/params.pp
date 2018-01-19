class prometheus::params {
  if defined('::tarmak') {
    $etcd_cluster         = $::tarmak::_etcd_cluster
    $etcd_k8s_main_port   = $::tarmak::etcd_k8s_main_client_port
    $etcd_k8s_events_port = $::tarmak::etcd_k8s_events_client_port
    $etcd_overlay_port    = $::tarmak::etcd_overlay_client_port
    $role                 = $::tarmak::role
  } else {
    $etcd_cluster         = undef
    $etcd_k8s_main_port   = 2379
    $etcd_k8s_events_port = 2369
    $etcd_overlay_port    = 2359
    $role                 = undef
  }
}
