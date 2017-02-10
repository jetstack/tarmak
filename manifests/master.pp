class puppernetes::master(
  $disable_kubelet = false,
  $disable_proxy = false,
){

  $apiserver_alt_names='kubernetes.default'
  $apiserver_ip_sans='10.254.0.1'
  include ::puppernetes
  include ::vault_client

  Class['vault_client'] -> Class['puppernetes::master']

  $service_account_key_path = "${::puppernetes::kubernetes_ssl_dir}/service-account-key.pem"
  vault_client::secret_service { 'service-account-key':
    field       => 'key',
    secret_path => "${::puppernetes::cluster_name}/secrets/service-accounts",
    user        => $::puppernetes::kubernetes_user,
    dest_path   => $service_account_key_path,
  }

  $controller_manager_base_path = "${::puppernetes::kubernetes_ssl_dir}/kube-controller-manager"
  vault_client::cert_service { 'kube-controller-manager':
    base_path   => $controller_manager_base_path,
    common_name => 'kube-controller-manager',
    role        => "${::puppernetes::cluster_name}/pki/${::puppernetes::kubernetes_ca_name}/sign/kube-controller-manager",
    user        => $::puppernetes::kubernetes_user,
  }

  $scheduler_base_path = "${::puppernetes::kubernetes_ssl_dir}/kube-scheduler"
  vault_client::cert_service { 'kube-scheduler':
    base_path   => $scheduler_base_path,
    common_name => 'kube-scheduler',
    role        => "${::puppernetes::cluster_name}/pki/${::puppernetes::kubernetes_ca_name}/sign/kube-scheduler",
    user        => $::puppernetes::kubernetes_user,
  }

  $apiserver_base_path = "${::puppernetes::kubernetes_ssl_dir}/kube-apiserver"
  vault_client::cert_service { 'kube-apiserver':
    base_path   => $apiserver_base_path,
    common_name => "kube-apiserver.${::puppernetes::cluster_name}.${::puppernetes::dns_root}",
    role        => "${::puppernetes::cluster_name}/pki/${::puppernetes::kubernetes_ca_name}/sign/kube-apiserver",
    user        => $::puppernetes::kubernetes_user,
    ip_sans     => "${apiserver_ip_sans},${::puppernetes::ipaddress}",
    alt_names   => $apiserver_alt_names,
  }

  $etcd_apiserver_base_path = "${::puppernetes::kubernetes_ssl_dir}/${::puppernetes::etcd_k8s_main_ca_name}"
  vault_client::cert_service { 'etcd-apiserver':
    base_path   => $etcd_apiserver_base_path,
    common_name => "${::hostname}.${::puppernetes::cluster_name}.${::puppernetes::dns_root}",
    role        => "${::puppernetes::cluster_name}/pki/${::puppernetes::etcd_k8s_main_ca_name}/sign/client",
    user        => $::puppernetes::kubernetes_user,
    ip_sans     => $::puppernetes::ipaddress,
  }

  class { 'kubernetes::apiserver':
      ca_file          => "${apiserver_base_path}-ca.pem",
      key_file         => "${apiserver_base_path}-key.pem",
      cert_file        => "${apiserver_base_path}.pem",
      etcd_ca_file     => "${etcd_apiserver_base_path}-ca.pem",
      etcd_key_file    => "${etcd_apiserver_base_path}-key.pem",
      etcd_cert_file   => "${etcd_apiserver_base_path}.pem",
      etcd_port        => $::puppernetes::etcd_k8s_main_client_port,
      etcd_events_port => $::puppernetes::etcd_k8s_events_client_port,
  }
  class { 'kubernetes::controller_manager':
      ca_file   => "${controller_manager_base_path}-ca.pem",
      key_file  => "${controller_manager_base_path}-key.pem",
      cert_file => "${controller_manager_base_path}.pem",
  }
  class { 'kubernetes::scheduler':
      ca_file   => "${scheduler_base_path}-ca.pem",
      key_file  => "${scheduler_base_path}-key.pem",
      cert_file => "${scheduler_base_path}.pem",
  }

  class { 'kubernetes::master':
    disable_kubelet => $disable_kubelet,
    disable_proxy   => $disable_proxy,
  }

}
