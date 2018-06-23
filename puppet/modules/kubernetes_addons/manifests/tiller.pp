class kubernetes_addons::tiller(
  Optional[String] $image=undef,
  Optional[String] $version=undef,
  String $namespace='kube-system',
) inherits ::kubernetes_addons::params {
  require ::kubernetes

  if $image == undef {
    $_image = 'gcr.io/kubernetes-helm/tiller'
  } else {
    $_image = $image
  }

  if $version == undef {
    $_version = '2.9.1'
  } else {
    $_version = $version
  }

  if versioncmp($::kubernetes::version, '1.6.0') >= 0 {
    $version_before_1_6 = false
  } else {
    $version_before_1_6 = true
  }

  if versioncmp($::kubernetes::version, '1.8.0') >= 0 {
    $version_before_1_8 = false
  } else {
    $version_before_1_8 = true
  }

  if versioncmp($::kubernetes::version, '1.9.0') >= 0 {
    $version_before_1_9 = false
  } else {
    $version_before_1_9 = true
  }

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
