# Tarmak
#
# This is the top-level class for the tarmak project. It's not including
# any component. It's just setting global variables for the cluster
#
# @example Declaring the class
#   include ::tarmak
# @example Overriding the kubernetes version
#   class{'tarmak':
#     kubernetes_version => '1.5.4'
#   }
#
# @param dest_dir path to installation directory for components
# @param bin_dir path to the binary directory for components
# @param cluster_name a DNS compatible name for the cluster
# @param systemctl_path absoulute path to systemctl binary
# @param role which role to build
# @param kubernetes_version Kubernetes version to install
# @param kubernetes_ca_name Name of the PKI resource in Vault for main Kubernetes CA
# @param kubernetes_api_proxy_ca_name Name of the PKI resource in Vault for the API server proxy
# @param kubernetes_api_aggregation Enable API aggregation for Kubernetes, defaults to true for versions 1.7+
class tarmak (
  String $dest_dir = '/opt',
  String $bin_dir = '/opt/bin',
  String $cluster_name = $tarmak::params::cluster_name,
  String $systemctl_path = $tarmak::params::systemctl_path,
  Enum['etcd','master','worker', nil] $role = nil,
  String $kubernetes_version = $tarmak::params::kubernetes_version,
  String $kubernetes_user = 'kubernetes',
  String $kubernetes_group = 'kubernetes',
  Integer $kubernetes_uid = 837,
  Integer $kubernetes_gid = 837,
  String $kubernetes_ca_name = 'k8s',
  String $kubernetes_api_proxy_ca_name = 'k8s-api-proxy',
  String $kubernetes_ssl_dir = '/etc/kubernetes/ssl',
  String $kubernetes_config_dir = '/etc/kubernetes',
  Optional[Boolean] $kubernetes_api_aggregation = undef,
  Optional[Boolean] $kubernetes_pod_security_policy = undef,
  Integer[1,65535] $kubernetes_api_insecure_port = 6443,
  Integer[1,65535] $kubernetes_api_secure_port = 8080,
  String $kubernetes_pod_network = '10.234.0.0/16',
  String $kubernetes_api_url = nil,
  String $kubernetes_api_prefix = 'api',
  Array[Enum['AlwaysAllow', 'ABAC', 'RBAC']] $kubernetes_authorization_mode = [],
  String $dns_root = $tarmak::params::dns_root,
  String $hostname = $tarmak::params::hostname,
  Array[String] $etcd_cluster = [],
  Integer[0,1] $etcd_start_index = 1,
  String $etcd_user = 'etcd',
  String $etcd_group = 'etcd',
  Integer $etcd_uid = 873,
  Integer $etcd_gid = 873,
  String $etcd_home = '/etc/etcd',
  String $etcd_ssl_dir = '/etc/etcd/ssl',
  Integer $etcd_instances = 3,
  String $etcd_advertise_client_network = $tarmak::params::etcd_advertise_client_network,
  Integer[1,65535] $etcd_overlay_client_port = 2359,
  Integer[1,65535] $etcd_overlay_peer_port = 2360,
  String $etcd_overlay_ca_name = 'etcd-overlay',
  String $etcd_overlay_version = '3.2.17',
  Integer[1,65535] $etcd_k8s_main_client_port = 2379,
  Integer[1,65535] $etcd_k8s_main_peer_port = 2380,
  String $etcd_k8s_main_ca_name = 'etcd-k8s',
  String $etcd_k8s_main_version = '3.2.17',
  Integer[1,65535] $etcd_k8s_events_client_port = 2369,
  Integer[1,65535] $etcd_k8s_events_peer_port = 2370,
  String $etcd_k8s_events_ca_name = 'etcd-k8s',
  String $etcd_k8s_events_version = '3.2.17',
  Enum['aws', ''] $cloud_provider = '',
  String $helper_path = $tarmak::params::helper_path,
  String $systemd_dir = '/etc/systemd/system',
) inherits ::tarmak::params {
  $ipaddress = $::ipaddress

  # decide if API aggregation should be enabled
  if $kubernetes_api_aggregation == undef {
    # enable after 1.7
    if versioncmp($kubernetes_version, '1.7.0') >= 0 {
      $_kubernetes_api_aggregation = true
    } else {
      $_kubernetes_api_aggregation = false
    }
  } else {
    $_kubernetes_api_aggregation = $kubernetes_api_aggregation
  }

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
    cluster_name        => $cluster_name,
    dns_root            => $dns_root,
    cloud_provider      => $cloud_provider,
    dest_dir            => $dest_dir,
    bin_dir             => $bin_dir,
    config_dir          => $kubernetes_config_dir,
    ssl_dir             => $kubernetes_ssl_dir,
    source              => 'gcs',
    master_url          => $_kubernetes_api_url,
    user                => $kubernetes_user,
    uid                 => $kubernetes_uid,
    group               => $kubernetes_group,
    gid                 => $kubernetes_gid,
    version             => $kubernetes_version,
    authorization_mode  => $kubernetes_authorization_mode,
    pod_network         => $kubernetes_pod_network,
    pod_security_policy => $kubernetes_pod_security_policy,
  }
}
