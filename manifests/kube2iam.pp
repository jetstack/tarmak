class kubernetes_addons::kube2iam(
  String $base_role_arn='',
  String $namespace='kube-system',
  String $image='jtblin/kube2iam',
  String $version='0.6.5',
  String $request_cpu='0.1',
  String $request_mem='64Mi',
  String $limit_cpu='',
  String $limit_mem='256Mi',
) {
  require ::kubernetes

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

  kubernetes::apply{'kube2iam':
    manifests => [
      template('kubernetes_addons/kube2iam-daemonset.yaml.erb'),
      template('kubernetes_addons/kube2iam-rbac.yaml.erb'),
    ],
  }
}
