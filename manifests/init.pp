# Puppernetes
#
# This is the top-level class for the puppernetes project. It's not including
# any component. It's just setting global variables for the cluster
#
# @example Declaring the class
#   include ::puppernetes
# @example Overriding the kubernetes version
#   class{'puppernetes':
#     kubernetes_version => '1.5.4'
#   }
#
# @param dest_dir path to installation directory for components
# @param bin_dir path to the binary directory for components
# @param cluster_name a DNS compatible name for the cluster
# @param vault_version vault version to use (deprecated)
# @param systemctl_path absoulute path to systemctl binary
# @param role which role to build
# @param kubernetes_version Kubernetes version to install
class puppernetes (
  String $dest_dir = '/opt',
  String $bin_dir = '/opt/bin',
  String $cluster_name = $puppernetes::params::cluster_name,
  String $systemctl_path = $puppernetes::params::systemctl_path,
  Enum['etcd','master','worker', nil] $role = nil,
  String $kubernetes_version = $puppernetes::params::kubernetes_version,
  String $kubernetes_user = 'kubernetes',
  String $kubernetes_group = 'kubernetes',
  Integer $kubernetes_uid = 837,
  Integer $kubernetes_gid = 837,
  String $kubernetes_ca_name = 'k8s',
  String $kubernetes_ssl_dir = '/etc/kubernetes/ssl',
  String $kubernetes_config_dir = '/etc/kubernetes',
  Integer[1,65535] $kubernetes_api_insecure_port = 6443,
  Integer[1,65535] $kubernetes_api_secure_port = 8080,
  String $kubernetes_pod_network = '10.234.0.0/16',
  String $kubernetes_api_url = nil,
  String $kubernetes_api_prefix = 'api',
  String $dns_root = $puppernetes::params::dns_root,
  String $hostname = $puppernetes::params::hostname,
  Array[String] $etcd_cluster = [],
  Integer[0,1] $etcd_start_index = 1,
  String $etcd_user = 'etcd',
  String $etcd_group = 'etcd',
  Integer $etcd_uid = 873,
  Integer $etcd_gid = 873,
  String $etcd_home = '/etc/etcd',
  String $etcd_ssl_dir = '/etc/etcd/ssl',
  Integer $etcd_instances = 3,
  String $etcd_advertise_client_network = $puppernetes::params::etcd_advertise_client_network,
  Integer[1,65535] $etcd_overlay_client_port = 2359,
  Integer[1,65535] $etcd_overlay_peer_port = 2360,
  String $etcd_overlay_ca_name = 'etcd-overlay',
  String $etcd_overlay_version = '2.3.7',
  Integer[1,65535] $etcd_k8s_main_client_port = 2379,
  Integer[1,65535] $etcd_k8s_main_peer_port = 2380,
  String $etcd_k8s_main_ca_name = 'etcd-k8s',
  String $etcd_k8s_main_version = '3.0.15',
  Integer[1,65535] $etcd_k8s_events_client_port = 2369,
  Integer[1,65535] $etcd_k8s_events_peer_port = 2370,
  String $etcd_k8s_events_ca_name = 'etcd-k8s',
  String $etcd_k8s_events_version = '3.0.15',
  Enum['aws', nil] $cloud_provider = nil,
  String $helper_path = $puppernetes::params::helper_path,
) inherits ::puppernetes::params {
  $ipaddress = $::ipaddress

  if $kubernetes_api_url == nil {
      $_kubernetes_api_url = "http://localhost:${$kubernetes_api_insecure_port}"
  }
  else {
      $_kubernetes_api_url = $kubernetes_api_url
  }

  if $etcd_cluster == [] {
    $_etcd_cluster = range($etcd_start_index, ($etcd_instances + $etcd_start_index -1)).map |$i| {
        "etcd-${i}.${cluster_name}.${dns_root}"
    }
  } else {
    $_etcd_cluster = $etcd_cluster
  }

  $kubernetes_pod_network_host = split($kubernetes_pod_network, '/')[0]
  $kubernetes_pod_network_mask = Integer(split($kubernetes_pod_network, '/')[1], 10)

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
