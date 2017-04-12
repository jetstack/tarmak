class calico::config{
  include ::kubernetes
  include ::calico

  $namespace = $::calico::namespace
  $etcd_endpoints = $::calico::etcd_endpoints

  $etcd_proto = $::calico::etcd_proto
  if $etcd_proto == 'https' {
    $etcd_tls_dir = $::calico::etcd_tls_dir
    $etcd_ca_file = $::calico::etcd_ca_file
    $etcd_cert_file = $::calico::etcd_cert_file
    $etcd_key_file = $::calico::etcd_key_file
  }

  kubernetes::apply{'calico-config':
    manifests => [
      template('calico/configmap.yaml.erb'),
    ],
  }
}
