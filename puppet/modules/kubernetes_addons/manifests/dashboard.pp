class kubernetes_addons::dashboard(
  String $image='gcr.io/google_containers/kubernetes-dashboard-amd64',
  Optional[String] $version=undef,
  String $limit_cpu='100m',
  String $limit_mem='128Mi',
  String $request_cpu='10m',
  String $request_mem='64Mi',
  $replicas=undef,
) inherits ::kubernetes_addons::params {
  require ::kubernetes

  if $version == undef {
    if versioncmp($::kubernetes::version, '1.8.0') >= 0 {
      $_version = '1.8.3'
    } elsif versioncmp($::kubernetes::version, '1.7.0') >= 0 {
      $_version = '1.7.1'
    } elsif versioncmp($::kubernetes::version, '1.6.0') >= 0 {
      $_version = '1.6.3'
    } elsif versioncmp($::kubernetes::version, '1.5.0') >= 0 {
      $_version = '1.5.1'
    } else {
      $_version = '1.4.2'
    }
  } else {
    $_version = $version
  }

  if versioncmp($_version, '1.7.0') >= 0 {
    $dashboard_version_before_1_7 = false
  } else {
    $dashboard_version_before_1_7 = true
  }

  if versioncmp($_version, '1.8.0') >= 0 {
    $dashboard_version_before_1_8 = false
  } else {
    $dashboard_version_before_1_8 = true
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

  kubernetes::apply{'kube-dashboard':
    manifests => [
      template('kubernetes_addons/dashboard-deployment.yaml.erb'),
      template('kubernetes_addons/dashboard-rbac.yaml.erb'),
    ],
  }
}
