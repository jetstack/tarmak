class tarmak::master(
  $disable_kubelet = true,
  $disable_proxy = true,
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
  ])
  $apiserver_ip_sans = unique([
    $::tarmak::ipaddress,
    '10.254.0.1',
    '127.0.0.1'
  ])

  Class['vault_client'] -> Class['tarmak::master']

  $service_account_key_path = "${::tarmak::kubernetes_ssl_dir}/service-account-key.pem"
  vault_client::secret_service { 'kube-service-account-key':
    field       => 'key',
    secret_path => "${::tarmak::cluster_name}/secrets/service-accounts",
    dest_path   => $service_account_key_path,
    uid         => $::tarmak::kubernetes_uid,
    gid         => $::tarmak::kubernetes_gid,
  }

  $controller_manager_base_path = "${::tarmak::kubernetes_ssl_dir}/kube-controller-manager"
  vault_client::cert_service { 'kube-controller-manager':
    base_path   => $controller_manager_base_path,
    common_name => 'system:kube-controller-manager',
    role        => "${::tarmak::cluster_name}/pki/${::tarmak::kubernetes_ca_name}/sign/kube-controller-manager",
    uid         => $::tarmak::kubernetes_uid,
    gid         => $::tarmak::kubernetes_gid,
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
    gid         => $::tarmak::kubernetes_gid,
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
    gid         => $::tarmak::kubernetes_gid,
    ip_sans     => $apiserver_ip_sans,
    alt_names   => $apiserver_alt_names,
    exec_post   => [
      "-${::tarmak::systemctl_path} --no-block try-restart kube-apiserver.service"
    ],
  }

  $admin_base_path = "${::tarmak::kubernetes_ssl_dir}/kube-admin"
  vault_client::cert_service { 'kube-admin':
    base_path   => $admin_base_path,
    common_name => 'admin',
    role        => "${::tarmak::cluster_name}/pki/${::tarmak::kubernetes_ca_name}/sign/admin",
    uid         => $::tarmak::kubernetes_uid,
    gid         => $::tarmak::kubernetes_gid,
  }

  $etcd_apiserver_base_path = "${::tarmak::kubernetes_ssl_dir}/${::tarmak::etcd_k8s_main_ca_name}"
  vault_client::cert_service { 'etcd-apiserver':
    base_path   => $etcd_apiserver_base_path,
    common_name => 'etcd-client',
    role        => "${::tarmak::cluster_name}/pki/${::tarmak::etcd_k8s_main_ca_name}/sign/client",
    ip_sans     => [$::tarmak::ipaddress],
    alt_names   => ["${::hostname}.${::tarmak::cluster_name}.${::tarmak::dns_root}"],
    uid         => $::tarmak::kubernetes_uid,
    gid         => $::tarmak::kubernetes_gid,
    exec_post   => [
      "-${::tarmak::systemctl_path} --no-block try-restart kube-apiserver.service"
    ],
  }

  class { 'kubernetes::apiserver':
      ca_file          => "${apiserver_base_path}-ca.pem",
      key_file         => "${apiserver_base_path}-key.pem",
      cert_file        => "${apiserver_base_path}.pem",
      etcd_ca_file     => "${etcd_apiserver_base_path}-ca.pem",
      etcd_key_file    => "${etcd_apiserver_base_path}-key.pem",
      etcd_cert_file   => "${etcd_apiserver_base_path}.pem",
      etcd_port        => $::tarmak::etcd_k8s_main_client_port,
      etcd_events_port => $::tarmak::etcd_k8s_events_client_port,
      etcd_nodes       => $::tarmak::_etcd_cluster,
      systemd_after    => ['kube-apiserver-cert.service', 'kube-service-account-key-secret.service'],
      systemd_requires => ['kube-apiserver-cert.service', 'kube-service-account-key-secret.service'],
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

  Service['kube-scheduler-cert.service'] -> Service['kube-scheduler.service']
  Service['kube-controller-manager-cert.service'] -> Service['kube-controller-manager.service']
  Service['kube-apiserver-cert.service'] -> Service['kube-apiserver.service']
  Service['kube-service-account-key-secret.service'] -> Service['kube-controller-manager.service']
  Service['kube-service-account-key-secret.service'] -> Service['kube-apiserver.service']
  Service['kube-admin-cert.service'] -> Kubernetes::Apply <||>

  class { 'kubernetes::master':
    disable_kubelet => $disable_kubelet,
    disable_proxy   => $disable_proxy,
  }

}
