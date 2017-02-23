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

  kubernetes::apply{'heapster':
    manifests => [
      template('kubernetes_addons/heapster-svc.yaml.erb'),
      template('kubernetes_addons/heapster-deployment.yaml.erb'),
    ],
  }
}
