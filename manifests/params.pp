class puppernetes::params{
  $cluster_name = 'cluster'
  $dns_root = 'jetstack.net'
  $hostname = $::hostname

  ## General
  $dest_dir = '/opt'
  $bin_dir = '/opt/bin'
  $helper_path = '/usr/local/sbin'
  $systemctl_path = $::osfamily ? {
    'RedHat' => '/usr/bin/systemctl',
    'Debian' => '/bin/systemctl',
    default  => '/usr/bin/systemctl',
  }

  ## Kubernetes
  $kubernetes_version = '1.5.2'
  $kubernetes_config_dir = '/etc/kubernetes'
  $kubernetes_ssl_dir = '/etc/kubernetes/ssl'
  $kubernetes_user = 'kubernetes'
  $kubernetes_group = 'kubernetes'
  $kubernetes_uid = 837
  $kubernetes_gid = 837
  $kubernetes_ca_name = 'k8s'
  $kubernetes_api_secure_port = 6443
  $kubernetes_api_insecure_port = 8080
  $kubernetes_pod_network = '10.234.0.0/16'
  $kubernetes_api_prefix = 'api'

  ## Etcd

  # etcd user + group
  $etcd_user = 'etcd'
  $etcd_group = 'etcd'
  $etcd_uid = 873
  $etcd_home = '/etc/etcd'
  $etcd_ssl_dir = '/etc/etcd/ssl'
  $etcd_instances = 3
  $etcd_start_index = 1
  $etcd_advertise_client_network = '172.16.0.0/12'

  # overlay etcd
  $etcd_overlay_client_port= 2359
  $etcd_overlay_peer_port = 2360
  $etcd_overlay_ca_name = 'etcd-overlay'
  $etcd_overlay_version = '2.3.7'

  # k8s etcds
  $etcd_k8s_main_client_port = 2379
  $etcd_k8s_main_peer_port = 2380
  $etcd_k8s_main_ca_name = 'etcd-k8s'
  $etcd_k8s_main_version = '3.0.15'
  $etcd_k8s_events_client_port = 2369
  $etcd_k8s_events_peer_port = 2370
  $etcd_k8s_events_ca_name = 'etcd-k8s'
  $etcd_k8s_events_version = '3.0.15'

  ## Vault
  $vault_version = '0.6.4'

}
