class calico::policy_controller (
  String $image = 'quay.io/calico/kube-controllers',
  String $version = '3.1.4',
)
{
  include ::kubernetes
  include ::calico

  $authorization_mode = $::kubernetes::_authorization_mode
  if member($authorization_mode, 'RBAC'){
    $rbac_enabled = true
  } else {
    $rbac_enabled = false
  }

  if versioncmp($::kubernetes::version, '1.6.0') >= 0 {
    $version_before_1_6 = false
  } else {
    $version_before_1_6 = true
  }

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
