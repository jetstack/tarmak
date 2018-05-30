class kubernetes_addons::tiller(
  String $image='gcr.io/kubernetes-helm/tiller',
  String $version='2.9.1',
  String $namespace='kube-system',
) inherits ::kubernetes_addons::params {
  require ::kubernetes

  $authorization_mode = $::kubernetes::_authorization_mode
  if member($authorization_mode, 'RBAC'){
    $rbac_enabled = true
  } else {
    $rbac_enabled = false
  }

  kubernetes::apply{'tiller':
    manifests => [
      template('kubernetes_addons/tiller-deployment.yaml.erb'),
    ],
  }
}
