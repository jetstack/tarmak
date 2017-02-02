class puppernetes::master {

  $apiserver_alt_names='kubernetes.default'
  $apiserver_ip_sans='10.254.0.1'

###
  #include ::puppernetes::node
#temporary
  user { 'k8s':
    ensure => present,
    uid    => $::puppernetes::k8s_uid,
    shell  => '/sbin/nologin',
    home   => $::puppernetes::k8s_home,
  }
###

  include ::vault_client

  #require ::vault_trust

  Class['vault_client'] -> Class['puppernetes::master']

###
  user { 'etcd':
    ensure => present,
    uid    => 873,
    shell  => '/sbin/nologin',
    home   => $::puppernetes::etcd_home,
  } ~>
  file { [ $::puppernetes::etcd_home, $::puppernetes::etcd_ssl_dir ]:
    ensure => directory,
    owner  => $::puppernetes::etcd_user,
    group  => $::puppernetes::etcd_group,
  }

  file { [ $::puppernetes::k8s_home, $::puppernetes::k8s_ssl_dir ]:
    ensure  => directory,
    owner   => $::puppernetes::k8s_user,
    group   => $::puppernetes::k8s_group,
    require => [ User[$::puppernetes::k8s_user], File[$::puppernetes::k8s_home] ],
  }
###

  vault_client::cert_service { 'controller-manager':
    base_path   => "${::puppernetes::k8s_ssl_dir}/kube-controller-manager",
    common_name => "kube-controller-manager.${::puppernetes::cluster_name}",
    role        => "${::puppernetes::cluster_name}/pki/k8s/sign/kube-controller-manager",
    user        =>  $::puppernetes::k8s_user,
    exec_post   =>  [ "${::puppernetes::helper_path}/helper read ${::puppernetes::cluster_name}/secrets/service-accounts key ${::puppernetes::k8s_ssl_dir}/service-account-key.pem",
                      "chown ${::puppernetes::k8s_user}:${::puppernetes::k8s_group} ${::puppernetes::k8s_ssl_dir/service-account-key.pem}",
                    ],
    require     => [User[$::puppernetes::k8s_user], Class['vault_client']]
  }

  vault_client::cert_service { 'scheduler':
    base_path   => "${::puppernetes::k8s_ssl_dir}/kube-scheduler",
    common_name => "kube-scheduler.${::puppernetes::cluster_name}",
    role        => "${::puppernetes::cluster_name}/pki/k8s/sign/kube-scheduler",
    user        => $::puppernetes::k8s_user,
    require     => [User[$::puppernetes::k8s_user], Class['vault_client']]
  }

  vault_client::cert_service { 'apiserver':
    base_path   => "${::puppernetes::k8s_ssl_dir}/kube-apiserver",
    common_name => "kube-apiserver.${::puppernetes::cluster_name}.${::puppernetes::dns_root}",
    role        => "${::puppernetes::cluster_name}/pki/k8s/sign/kube-apiserver",
    user        => $::puppernetes::k8s_user,
    ip_sans     => "${apiserver_ip_sans},${::puppernetes::ipaddress}",
    alt_names   => $apiserver_alt_names,
    exec_post   =>  [ "${::puppernetes::helper_path}/helper read ${::puppernetes::cluster_name}/secrets/service-accounts key ${::puppernetes::k8s_ssl_dir}/service-account-key.pem",
                        'chown ${::puppernetes::k8s_user}:${::puppernetes::k8s_group} ${::puppernetes::k8s_ssl_dir}/service-account-key.pem',
                      ],
    require     => [User[$::puppernetes::k8s_user], Class['vault_client']],
  }

  vault_client::cert_service { 'etcd-k8s':
    base_path   => "${::puppernetes::etcd_ssl_path}/etcd-${::puppernetes::etcd_k8s_main_ca_name}",
    common_name => "${::hostname}.${::puppernetes::cluster_name}.${::puppernetes::dns_root}",
    role        => "${::puppernetes::cluster_name}/pki/etcd-k8s/sign/client",
    user        => $::puppernetes::k8s_user,
    ip_sans     => $::puppernetes::ipaddress,
    require     => [ User[$::puppernetes::k8s_user,], File[$::puppernetes::etcd_ssl_dir], Class['vault_client'] ]
  }
}
