class kubernetes_addons::heapster(
  $image=$::kubernetes_addons::params::heapster_image,
  $version=$::kubernetes_addons::params::heapster_version,
  $cpu=$::kubernetes_addons::params::heapster_cpu,
  $mem=$::kubernetes_addons::params::heapster_mem,
  $extra_cpu=$::kubernetes_addons::params::heapster_extra_cpu,
  $extra_mem=$::kubernetes_addons::params::heapster_extra_mem,
  $nanny_image=$::kubernetes_addons::params::heapster_nanny_image,
  $nanny_version=$::kubernetes_addons::params::heapster_nanny_version,
  $nanny_request_cpu=$::kubernetes_addons::params::heapster_nanny_request_cpu,
  $nanny_request_mem=$::kubernetes_addons::params::heapster_nanny_request_mem,
  $nanny_limit_cpu=$::kubernetes_addons::params::heapster_nanny_limit_cpu,
  $nanny_limit_mem=$::kubernetes_addons::params::heapster_nanny_limit_mem,
  $sink=undef,
) inherits ::kubernetes_addons::params {
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

  kubernetes::apply{'heapster':
    manifests => [
      template('kubernetes_addons/heapster-svc.yaml.erb'),
      template('kubernetes_addons/heapster-deployment.yaml.erb'),
      template('kubernetes_addons/heapster-rbac.yaml.erb'),
    ],
  }
}
