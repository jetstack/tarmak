class kubernetes_addons::cluster_autoscaler(
  String $image='gcr.io/google_containers/cluster-autoscaler',
  String $version='',
  String $limit_cpu='200m',
  String $limit_mem='500Mi',
  String $request_cpu='100m',
  String $request_mem='300Mi',
  Array[String] $instance_pool_names=[],
  Array[Integer] $min_instances=[],
  Array[Integer] $max_instances=[],
  Optional[Boolean] $enable_overprovisioning=undef,
  Optional[String] $proportional_image=undef,
  Optional[String] $proportional_version=undef,
  Optional[Float] $scale_down_utilization_threshold = undef,
  Integer $reserved_millicores_per_replica = 0,
  Integer $reserved_megabytes_per_replica = 0,
  Integer $cores_per_replica = 0,
  Integer $nodes_per_replica = 0,
  Integer $replica_count = 0,
  $ca_mounts=$::kubernetes_addons::params::ca_mounts,
  $cloud_provider=$::kubernetes_addons::params::cloud_provider,
  $aws_region=$::kubernetes_addons::params::aws_region,
  Enum['present', 'absent'] $ensure = 'present',
) inherits ::kubernetes_addons::params {
  require ::kubernetes

  $authorization_mode = $::kubernetes::_authorization_mode
  if member($authorization_mode, 'RBAC'){
    $rbac_enabled = true
  } else {
    $rbac_enabled = false
  }

  if defined('$kubernetes::cluster_name') {
    $asg_name_prefix="${::kubernetes::cluster_name}-kubernetes-"
  } else {
    fail('cluster name must be defined')
  }

  if $version == '' {
    if versioncmp($::kubernetes::version, '1.11.0') >= 0 {
      $_version = '1.3.0'
    } elsif versioncmp($::kubernetes::version, '1.10.0') >= 0 {
      $_version = '1.2.2'
    } elsif versioncmp($::kubernetes::version, '1.9.0') >= 0 {
      $_version = '1.1.2'
    } elsif versioncmp($::kubernetes::version, '1.8.0') >= 0 {
      $_version = '1.0.4'
    } elsif versioncmp($::kubernetes::version, '1.7.0') >= 0 {
      $_version = '0.6.4'
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

  if versioncmp($_version, '0.6.0') >= 0 {
    $balance_similar_node_groups = true
  }

  if versioncmp($::kubernetes::version, '1.6.0') >= 0 {
    $version_before_1_6 = false
  } else {
    $version_before_1_6 = true
  }

  if $cores_per_replica == 0 and $nodes_per_replica == 0 {
    if $replica_count == 0 {
      $_replica_count = 1
    } else {
      $_replica_count = $replica_count
    }
  }

  if $enable_overprovisioning == undef {
    $_enable_overprovisioning = false
  } else {
    $_enable_overprovisioning = $enable_overprovisioning
  }

  if $proportional_version == undef {
    $_proportional_version = '1.1.2'
  } else {
    $_proportional_version = $proportional_version
  }

  if $proportional_image == undef {
    $_proportional_image = 'k8s.gcr.io/cluster-proportional-autoscaler-amd64'
  } else {
    $_proportional_image = $proportional_image
  }


  if $_enable_overprovisioning and versioncmp($::kubernetes::version, '1.9.0') >= 0 {
    $overprovision_ensure = $ensure
  } else {
    $overprovision_ensure = 'absent'
  }

  kubernetes::apply{'cluster-autoscaler-overprovisioning':
    ensure    => $overprovision_ensure,
    manifests => [
      template('kubernetes_addons/cluster-autoscaler-overprovisioning.yaml.erb'),
      template('kubernetes_addons/cluster-autoscaler-overprovisioning-rbac.yaml.erb'),
    ],
  }

  kubernetes::apply{'cluster-autoscaler':
    ensure    => $ensure,
    manifests => [
      template('kubernetes_addons/cluster-autoscaler-deployment.yaml.erb'),
      template('kubernetes_addons/cluster-autoscaler-rbac.yaml.erb'),
    ],
  }
}
