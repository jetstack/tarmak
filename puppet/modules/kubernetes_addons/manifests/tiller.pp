class kubernetes_addons::tiller(
  String $image='gcr.io/kubernetes-helm/tiller',
  String $version='2.9.1',
  String $namespace='kube-system',
  Enum['present', 'absent'] $ensure = 'present'
) inherits ::kubernetes_addons::params {
  require ::kubernetes

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
    ensure    => $ensure,
    manifests => [
      template('kubernetes_addons/tiller-deployment.yaml.erb'),
    ],
  }
}
