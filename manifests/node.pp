class calico::node (
  String $node_image = 'quay.io/calico/node',
  String $node_version = '1.1.1',
  String $cni_image = 'quay.io/calico/cni',
  String $cni_version = '1.6.2',
  String $ipv4_pool_cidr = '10.231.0.0/16',
  Enum['always', 'cross-subnet', 'off'] $ipv4_pool_ipip_mode = 'always',
)
{
  include ::kubernetes
  include ::calico

  $namespace = $::calico::namespace
  $etcd_cert_path = $::calico::etcd_cert_path

  kubernetes::apply{'calico-node':
    manifests => [
      template('calico/node-daemonset.yaml.erb'),
    ],
  }

}
