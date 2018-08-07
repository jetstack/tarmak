class tarmak::master(
  $disable_kubelet = true,
  $disable_proxy = true,
  Array[String] $apiserver_additional_san_domains = [],
  Array[String] $apiserver_additional_san_ips = [],
){
  include ::tarmak
  include ::vault_client

  $apiserver_alt_names = unique([
    "${::tarmak::kubernetes_api_prefix}.${::tarmak::cluster_name}.${::tarmak::dns_root}",
    'kubernetes',
    'kubernetes.default',
    'kubernetes.default.svc',
    'kubernetes.default.svc.cluster.local',
    'localhost'
  ] + $apiserver_additional_san_domains)
  $apiserver_ip_sans = unique([
    $::tarmak::ipaddress,
    '10.254.0.1',
    '127.0.0.1'
  ] + $apiserver_additional_san_ips)

  Class['vault_client'] -> Class['tarmak::master']

  $service_account_key_path = "${::tarmak::kubernetes_ssl_dir}/service-account-key.pem"
  vault_client::secret_service { 'kube-service-account-key':
    field       => 'key',
    secret_path => "${::tarmak::cluster_name}/secrets/service-accounts",
    dest_path   => $service_account_key_path,
    uid         => $::tarmak::kubernetes_uid,
  }

  $encryption_config_file = "${::tarmak::kubernetes_config_dir}/encryption-config.yaml"
  vault_client::secret_service { 'kube-encryption-config-file':
    field       => 'content',
    secret_path => "${::tarmak::cluster_name}/secrets/encryption-config",
    dest_path   => $encryption_config_file,
    uid         => $::tarmak::kubernetes_uid,
  }

  $controller_manager_base_path = "${::tarmak::kubernetes_ssl_dir}/kube-controller-manager"
  vault_client::cert_service { 'kube-controller-manager':
    base_path   => $controller_manager_base_path,
    common_name => 'system:kube-controller-manager',
    role        => "${::tarmak::cluster_name}/pki/${::tarmak::kubernetes_ca_name}/sign/kube-controller-manager",
    uid         => $::tarmak::kubernetes_uid,
    exec_post   => [
      "-${::tarmak::systemctl_path} --no-block try-restart kube-controller-manager.service"
    ],
  }

  $scheduler_base_path = "${::tarmak::kubernetes_ssl_dir}/kube-scheduler"
  vault_client::cert_service { 'kube-scheduler':
    base_path   => $scheduler_base_path,
    common_name => 'system:kube-scheduler',
    role        => "${::tarmak::cluster_name}/pki/${::tarmak::kubernetes_ca_name}/sign/kube-scheduler",
    uid         => $::tarmak::kubernetes_uid,
    exec_post   => [
      "-${::tarmak::systemctl_path} --no-block try-restart kube-scheduler.service"
    ],
  }

  $apiserver_base_path = "${::tarmak::kubernetes_ssl_dir}/kube-apiserver"
  vault_client::cert_service { 'kube-apiserver':
    base_path   => $apiserver_base_path,
    common_name => 'system:kube-apiserver',
    role        => "${::tarmak::cluster_name}/pki/${::tarmak::kubernetes_ca_name}/sign/kube-apiserver",
    uid         => $::tarmak::kubernetes_uid,
    ip_sans     => $apiserver_ip_sans,
    alt_names   => $apiserver_alt_names,
    exec_post   => [
      "-${::tarmak::systemctl_path} --no-block try-restart kube-apiserver.service"
    ],
  }

  $apiserver_dependencies_base =  [
    'kube-apiserver-cert.service',
    'kube-admin-cert.service',
    'kube-service-account-key-secret.service'
  ]

  if $::tarmak::_kubernetes_api_aggregation {
    $apiserver_proxy_base_path = "${::tarmak::kubernetes_ssl_dir}/kube-apiserver-proxy"
    vault_client::cert_service { 'kube-apiserver-proxy':
      base_path   => $apiserver_proxy_base_path,
      common_name => 'kube-apiserver-proxy',
      role        => "${::tarmak::cluster_name}/pki/${::tarmak::kubernetes_api_proxy_ca_name}/sign/kube-apiserver",
      uid         => $::tarmak::kubernetes_uid,
      exec_post   => [
        "-${::tarmak::systemctl_path} --no-block try-restart kube-apiserver.service"
      ],
    }
    $apiserver_dependencies = $apiserver_dependencies_base + 'kube-apiserver-proxy-cert.service'
    $requestheader_client_ca_file = "${apiserver_proxy_base_path}-ca.pem"
    $proxy_client_cert_file = "${apiserver_proxy_base_path}.pem"
    $proxy_client_key_file = "${apiserver_proxy_base_path}-key.pem"
  } else {
    $apiserver_dependencies = $apiserver_dependencies_base
    $requestheader_client_ca_file = undef
    $proxy_client_cert_file = undef
    $proxy_client_key_file = undef
  }

  $admin_base_path = "${::tarmak::kubernetes_ssl_dir}/kube-admin"
  vault_client::cert_service { 'kube-admin':
    base_path   => $admin_base_path,
    common_name => 'admin',
    role        => "${::tarmak::cluster_name}/pki/${::tarmak::kubernetes_ca_name}/sign/admin",
    uid         => $::tarmak::kubernetes_uid,
  }

  $etcd_apiserver_base_path = "${::tarmak::kubernetes_ssl_dir}/${::tarmak::etcd_k8s_main_ca_name}"
  vault_client::cert_service { 'etcd-apiserver':
    base_path   => $etcd_apiserver_base_path,
    common_name => 'etcd-client',
    role        => "${::tarmak::cluster_name}/pki/${::tarmak::etcd_k8s_main_ca_name}/sign/client",
    ip_sans     => [$::tarmak::ipaddress],
    alt_names   => ["${::hostname}.${::tarmak::cluster_name}.${::tarmak::dns_root}"],
    uid         => $::tarmak::kubernetes_uid,
    exec_post   => [
      "-${::tarmak::systemctl_path} --no-block try-restart kube-apiserver.service"
    ],
  }

  class { 'kubernetes::apiserver':
      ca_file                      => "${apiserver_base_path}-ca.pem",
      key_file                     => "${apiserver_base_path}-key.pem",
      cert_file                    => "${apiserver_base_path}.pem",
      etcd_ca_file                 => "${etcd_apiserver_base_path}-ca.pem",
      etcd_key_file                => "${etcd_apiserver_base_path}-key.pem",
      etcd_cert_file               => "${etcd_apiserver_base_path}.pem",
      etcd_port                    => $::tarmak::etcd_k8s_main_client_port,
      etcd_events_port             => $::tarmak::etcd_k8s_events_client_port,
      etcd_nodes                   => $::tarmak::_etcd_cluster,
      kubelet_client_key_file      => "${admin_base_path}-key.pem",
      kubelet_client_cert_file     => "${admin_base_path}.pem",
      systemd_after                => $apiserver_dependencies,
      systemd_requires             => $apiserver_dependencies,
      requestheader_client_ca_file => $requestheader_client_ca_file,
      proxy_client_cert_file       => $proxy_client_cert_file ,
      proxy_client_key_file        => $proxy_client_key_file,
      encryption_config_file       => $encryption_config_file,
  }

  class { 'kubernetes::controller_manager':
      ca_file          => "${controller_manager_base_path}-ca.pem",
      key_file         => "${controller_manager_base_path}-key.pem",
      cert_file        => "${controller_manager_base_path}.pem",
      systemd_after    => ['kube-controller-manager-cert.service', 'kube-service-account-key-secret.service'],
      systemd_requires => ['kube-controller-manager-cert.service', 'kube-service-account-key-secret.service'],
  }

  class { 'kubernetes::scheduler':
      ca_file          => "${scheduler_base_path}-ca.pem",
      key_file         => "${scheduler_base_path}-key.pem",
      cert_file        => "${scheduler_base_path}.pem",
      systemd_after    => ['kube-scheduler-cert.service'],
      systemd_requires => ['kube-scheduler-cert.service'],
  }

  class { 'kubernetes::kubectl':
      ca_file   => "${admin_base_path}-ca.pem",
      key_file  => "${admin_base_path}-key.pem",
      cert_file => "${admin_base_path}.pem",
  }

  Vault_Client::Cert_Service['kube-scheduler'] -> Service['kube-scheduler.service']
  Vault_Client::Cert_Service['kube-controller-manager'] -> Service['kube-controller-manager.service']
  Vault_Client::Cert_Service['kube-apiserver'] -> Service['kube-apiserver.service']
  Vault_Client::Cert_Service['kube-admin'] -> Service['kube-apiserver.service']
  Vault_Client::Secret_Service['kube-service-account-key'] -> Service['kube-controller-manager.service']
  Vault_Client::Secret_Service['kube-service-account-key'] -> Service['kube-apiserver.service']
  Service['kube-admin-cert.service'] -> Kubernetes::Apply <||>

  class { 'kubernetes::master':
    disable_kubelet => $disable_kubelet,
    disable_proxy   => $disable_proxy,
  }

}
