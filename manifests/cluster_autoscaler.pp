class kubernetes_addons::cluster_autoscaler(
  $image=$::kubernetes_addons::params::cluster_autoscaler_image,
  String $version='',
  $request_cpu=$::kubernetes_addons::params::cluster_autoscaler_request_cpu,
  $request_mem=$::kubernetes_addons::params::cluster_autoscaler_request_mem,
  $limit_cpu=$::kubernetes_addons::params::cluster_autoscaler_limit_cpu,
  $limit_mem=$::kubernetes_addons::params::cluster_autoscaler_limit_mem,
  $asg_name=$::kubernetes_addons::params::cluster_autoscaler_asg_name,
  $min_instances=$::kubernetes_addons::params::cluster_autoscaler_min_instances,
  $max_instances=$::kubernetes_addons::params::cluster_autoscaler_max_instances,
  $ca_mounts=$::kubernetes_addons::params::ca_mounts,
  $cloud_provider=$::kubernetes_addons::params::cloud_provider,
  $aws_region=$::kubernetes_addons::params::aws_region,
) inherits ::kubernetes_addons::params {
  require ::kubernetes

  $authorization_mode = $::kubernetes::_authorization_mode
  if member($authorization_mode, 'RBAC'){
    $rbac_enabled = true
  } else {
    $rbac_enabled = false
  }

  if $version == '' {
    if versioncmp($::kubernetes::version, '1.7.0') >= 0 {
      $_version = '0.6.0'
    } elsif versioncmp($::kubernetes::version, '1.6.0') >= 0 {
      $_version = '0.5.4'
    } elsif versioncmp($::kubernetes::version, '1.5.0') >= 0 {
      $_version = '0.4.0'
    } else {
      $_version = '0.3.0'
    }
  } else {
    $_version = $version
  }

  if versioncmp($::kubernetes::version, '1.6.0') >= 0 {
    $version_before_1_6 = false
  } else {
    $version_before_1_6 = true
  }


  kubernetes::apply{'cluster-autoscaler':
    manifests => [
      template('kubernetes_addons/cluster-autoscaler-deployment.yaml.erb'),
    ],
  }
}
