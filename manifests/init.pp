# Class: puppernetes
class puppernetes (
  String $cluster_name = $puppernetes::params::cluster_name,
  String $dns_root = $puppernetes::params::dns_root,
  String $kubernetes_version = $puppernetes::params::kubernetes_version,
  String $etcd_user = $puppernetes::params::etcd_user,
  String $etcd_group = $puppernetes::params::etcd_group,
  String $etcd_config_dir = $puppernetes::params::etcd_config_dir,
  String $etcd_ssl_dir = $puppernetes::params::etcd_ssl_dir,
  Integer $etcd_instances = $puppernetes::params::etcd_instances,
  String $etcd_advertise_client_network = $puppernetes::params::etcd_advertise_client_network,
  Integer $etcd_overlay_client_port = $puppernetes::params::etcd_overlay_client_port,
  Integer $etcd_overlay_peer_port = $puppernetes::params::etcd_overlay_peer_port,
  String $etcd_overlay_ca_name = $puppernetes::params::etcd_overlay_ca_name,
  String $etcd_overlay_version = $puppernetes::params::etcd_overlay_version,
  Integer $etcd_k8s_main_client_port = $puppernetes::params::etcd_k8s_main_client_port,
  Integer $etcd_k8s_main_peer_port = $puppernetes::params::etcd_k8s_main_peer_port,
  String $etcd_k8s_main_ca_name = $puppernetes::params::etcd_k8s_main_ca_name,
  String $etcd_k8s_main_version = $puppernetes::params::etcd_k8s_main_version,
  Integer $etcd_k8s_events_client_port = $puppernetes::params::etcd_k8s_events_client_port,
  Integer $etcd_k8s_events_peer_port = $puppernetes::params::etcd_k8s_events_peer_port,
  String $etcd_k8s_events_ca_name = $puppernetes::params::etcd_k8s_events_ca_name,
  String $etcd_k8s_events_version = $puppernetes::params::etcd_k8s_events_version,
  String $vault_version = $puppernetes::params::vault_version,
  String $cloud_provider = $puppernetes::params::cloud_provider,
) inherits ::puppernetes::params {

}
