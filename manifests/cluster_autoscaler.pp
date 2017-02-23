class kubernetes_addons::cluster_autoscaler(
  $image=$::kubernetes_addons::params::cluster_autoscaler_image,
  $version=$::kubernetes_addons::params::cluster_autoscaler_version,
  $request_cpu=$::kubernetes_addons::params::cluster_autoscaler_request_cpu,
  $request_mem=$::kubernetes_addons::params::cluster_autoscaler_request_mem,
  $limit_cpu=$::kubernetes_addons::params::cluster_autoscaler_limit_cpu,
  $limit_mem=$::kubernetes_addons::params::cluster_autoscaler_limit_mem,
  $asg_name=$::kubernetes_addons::params::cluster_autoscaler_asg_name,
  $min_instances=$::kubernetes_addons::params::cluster_autoscaler_min_instances,
  $max_instances=$::kubernetes_addons::params::cluster_autoscaler_max_instances,
  $ca_bundle_path=$::kubernetes_addons::params::ca_bundle_path,
  $cloud_provider=$::kubernetes_addons::params::cloud_provider,
) inherits ::kubernetes_addons::params {
  require ::kubernetes

  kubernetes::apply{'cluster-autoscaler':
    manifests => [
      template('kubernetes_addons/cluster-autoscaler-deployment.yaml.erb'),
    ],
  }
}
