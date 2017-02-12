# Class: puppernetes
class puppernetes (
  $dest_dir = $::puppernetes::params::dest_dir,
  $bin_dir = $::puppernetes::params::bin_dir,
  $cluster_name = $puppernetes::params::cluster_name,
  $vault_version = $puppernetes::params::vault_version,
  $kubernetes_version = $puppernetes::params::kubernetes_version,
  $kubernetes_user = $puppernetes::params::kubernetes_user,
  $kubernetes_group = $puppernetes::params::kubernetes_group,
  $kubernetes_uid = $puppernetes::params::kubernetes_uid,
  $kubernetes_gid = $puppernetes::params::kubernetes_gid,
  $kubernetes_ca_name = $puppernetes::params::kubernetes_ca_name,
  $kubernetes_ssl_dir = $puppernetes::params::kubernetes_ssl_dir,
  $kubernetes_config_dir = $puppernetes::params::kubernetes_config_dir,
  $kubernetes_api_insecure_port = $puppernetes::params::kubernetes_api_insecure_port,
  $kubernetes_api_secure_port = $puppernetes::params::kubernetes_api_secure_port,
  $kubernetes_api_url = undef,
  $dns_root = $puppernetes::params::dns_root,
  $etcd_user = $puppernetes::params::etcd_user,
  $etcd_group = $puppernetes::params::etcd_group,
  $etcd_uid = $puppernetes::params::etcd_uid,
  $etcd_home = $puppernetes::params::etcd_home,
  $etcd_ssl_dir = $puppernetes::params::etcd_ssl_dir,
  $etcd_instances = $puppernetes::params::etcd_instances,
  $etcd_advertise_client_network = $puppernetes::params::etcd_advertise_client_network,
  $etcd_overlay_client_port = $puppernetes::params::etcd_overlay_client_port,
  $etcd_overlay_peer_port = $puppernetes::params::etcd_overlay_peer_port,
  $etcd_overlay_ca_name = $puppernetes::params::etcd_overlay_ca_name,
  $etcd_overlay_version = $puppernetes::params::etcd_overlay_version,
  $etcd_k8s_main_client_port = $puppernetes::params::etcd_k8s_main_client_port,
  $etcd_k8s_main_peer_port = $puppernetes::params::etcd_k8s_main_peer_port,
  $etcd_k8s_main_ca_name = $puppernetes::params::etcd_k8s_main_ca_name,
  $etcd_k8s_main_version = $puppernetes::params::etcd_k8s_main_version,
  $etcd_k8s_events_client_port = $puppernetes::params::etcd_k8s_events_client_port,
  $etcd_k8s_events_peer_port = $puppernetes::params::etcd_k8s_events_peer_port,
  $etcd_k8s_events_ca_name = $puppernetes::params::etcd_k8s_events_ca_name,
  $etcd_k8s_events_version = $puppernetes::params::etcd_k8s_events_version,
  $cloud_provider = undef,
  $helper_path = $puppernetes::params::helper_path,
) inherits ::puppernetes::params {
  $ipaddress = $::ipaddress

  if $kubernetes_api_url == undef {
      $_kubernetes_api_url = "http://localhost:${$kubernetes_api_insecure_port}"
  }
  else {
      $_kubernetes_api_url = $kubernetes_api_url
  }

  class { 'kubernetes':
    cluster_name   => $cluster_name,
    dns_root       => $dns_root,
    cloud_provider => $cloud_provider,
    dest_dir       => $dest_dir,
    bin_dir        => $bin_dir,
    config_dir     => $kubernetes_config_dir,
    ssl_dir        => $kubernetes_ssl_dir,
    source         => 'gcs',
    master_url     => $_kubernetes_api_url,
    user           => $kubernetes_user,
    uid            => $kubernetes_uid,
    group          => $kubernetes_group,
    gid            => $kubernetes_gid,
    version        => $kubernetes_version,
  }
}
