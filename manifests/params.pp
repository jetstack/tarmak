class puppernetes::params{
  $cluster_name = 'cluster'
  $dns_root = 'jetstack.net'

  ## Kubernetes
  $kubernetes_version = '1.5.1'

  ## Etcd

  # etcd user + group
  $etcd_user = 'etcd'
  $etcd_group = 'etcd'
  $etcd_config_dir = '/etc/etcd'
  $etcd_ssl_dir = '/etc/etcd/ssl'
  $etcd_instances = 3
  $etcd_advertise_client_network = '172.16.0.0/12'

  # overlay etcd
  $etcd_overlay_client_port= 2359
  $etcd_overlay_peer_port = 2360
  $etcd_overlay_ca_name = 'overlay'
  $etcd_overlay_version = '2.3.7'

  # k8s etcds
  $etcd_k8s_main_client_port = 2379
  $etcd_k8s_main_peer_port = 2380
  $etcd_k8s_main_ca_name = 'k8s'
  $etcd_k8s_main_version = '3.0.15'
  $etcd_k8s_events_client_port = 2369
  $etcd_k8s_events_peer_port = 2370
  $etcd_k8s_events_ca_name = 'k8s'
  $etcd_k8s_events_version = '3.0.15'

  ## Vault
  $vault_version = '0.6.4'

  ## Cloud Provider
  $cloud_provider = 'vagrant'

}
