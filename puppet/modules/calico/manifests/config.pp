class calico::config{
  include ::kubernetes
  include ::calico

  $mtu = $::calico::mtu
  $namespace = $::calico::namespace

  if $::calico::backend == 'etcd' {
    $etcd_endpoints = $::calico::etcd_endpoints
    $etcd_proto = $::calico::etcd_proto

    if $etcd_proto == 'https' {
      $etcd_tls_dir = $::calico::etcd_tls_dir
      $etcd_ca_file = $::calico::etcd_ca_file
      $etcd_cert_file = $::calico::etcd_cert_file
      $etcd_key_file = $::calico::etcd_key_file
    }

    kubernetes::apply{'calico-config':
      ensure    => 'present',
      manifests => [
        template('calico/configmap_etcd.yaml.erb'),
      ],
    }

  } else {
    $pod_network = $::calico::pod_network

    kubernetes::apply{'calico-config':
      ensure    => 'present',
      manifests => [
        template('calico/configmap_kubernetes.yaml.erb'),
      ],
    }
  }
}
