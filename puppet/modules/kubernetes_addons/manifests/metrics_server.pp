class kubernetes_addons::metrics_server(
  Optional[String] $version=undef,
  String $image='gcr.io/google_containers/metrics-server-amd64',
  String $cpu='40m',
  String $mem='100Mi',
  String $extra_cpu='0.5m',
  String $extra_mem='4Mi',
  String $nanny_version='1.8.1',
  String $nanny_image='gcr.io/google_containers/addon-resizer',
  String $nanny_request_cpu='5m',
  String $nanny_request_mem='50Mi',
  String $nanny_limit_cpu='100m',
  String $nanny_limit_mem='300Mi',
) inherits ::kubernetes_addons::params {
  require ::kubernetes

  if $version == undef {
    if versioncmp($::kubernetes::version, '1.8.0') >= 0 {
      $_version = '0.3.0'
    } else {
      $_version = '0.1.0'
    }
  } else {
    $_version = $version
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

  if versioncmp($::kubernetes::version, '1.7.0') >= 0 {
    kubernetes::apply{'metrics-server':
      manifests => [
        template('kubernetes_addons/metrics-server.yaml.erb'),
        template('kubernetes_addons/metrics-server-rbac.yaml.erb'),
      ],
    }
  }
}
