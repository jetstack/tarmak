class calico::policy_controller (
  String $image = 'quay.io/calico/kube-policy-controller',
  String $version = '0.5.4',
)
{
  include ::kubernetes
  include ::calico

  if $::calico::backend == 'etcd' {
    $namespace = $::calico::namespace

    if $::calico::etcd_proto == 'https' {
      $etcd_tls_dir = $::calico::etcd_tls_dir
      $tls = true
    } else {
      $tls = false
    }


    kubernetes::apply{'calico-policy-controller':
      manifests => [
        template('calico/policy-controller-deployment.yaml.erb'),
      ],
    }
  }
}
