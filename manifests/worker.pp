class puppernetes::worker {
  include ::puppernetes
  require ::vault_client

  $proxy_base_path = "${::puppernetes::kubernetes_ssl_dir}/kube-proxy"
  vault_client::cert_service { 'kube-proxy':
    base_path   => $proxy_base_path,
    common_name => 'kube-proxy',
    role        => "${::puppernetes::cluster_name}/pki/${::puppernetes::kubernetes_ca_name}/sign/kube-proxy",
    user        => $::puppernetes::kubernetes_user,
    require     => [
      User[$::puppernetes::kubernetes_user],
      Class['vault_client']
    ],
    exec_post   => [
      "-${::puppernetes::systemctl_path} --no-block try-restart kube-proxy.service"
    ],
  }

  $kubelet_base_path = "${::puppernetes::kubernetes_ssl_dir}/kubelet"
  vault_client::cert_service { 'kubelet':
    base_path   => $kubelet_base_path,
    common_name => 'kubelet',
    role        => "${::puppernetes::cluster_name}/pki/${::puppernetes::kubernetes_ca_name}/sign/kubelet",
    user        => $::puppernetes::kubernetes_user,
    require     => [
      User[$::puppernetes::kubernetes_user],
      Class['vault_client']
    ],
    exec_post   => [
      "-${::puppernetes::systemctl_path} --no-block try-restart kubelet.service"
    ],
  }

  class { 'kubernetes::kubelet':
      ca_file   => "${kubelet_base_path}-ca.pem",
      key_file  => "${kubelet_base_path}-key.pem",
      cert_file => "${kubelet_base_path}.pem",
  }

  class { 'kubernetes::proxy':
      ca_file   => "${proxy_base_path}-ca.pem",
      key_file  => "${proxy_base_path}-key.pem",
      cert_file => "${proxy_base_path}.pem",
  }

  class { 'kubernetes::worker':

  }

}
