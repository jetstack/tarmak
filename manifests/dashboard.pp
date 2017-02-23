class kubernetes_addons::dashboard(
  $image=$::kubernetes_addons::params::dashboard_image,
  $version=$::kubernetes_addons::params::dashboard_version,
  $request_cpu=$::kubernetes_addons::params::dashboard_request_cpu,
  $request_mem=$::kubernetes_addons::params::dashboard_request_mem,
  $limit_cpu=$::kubernetes_addons::params::dashboard_limit_cpu,
  $limit_mem=$::kubernetes_addons::params::dashboard_limit_mem,
) inherits ::kubernetes_addons::params {
  require ::kubernetes

  kubernetes::apply{'kube-dashboard':
    manifests => [
      template('kubernetes_addons/dashboard-svc.yaml.erb'),
      template('kubernetes_addons/dashboard-deployment.yaml.erb'),
    ],
  }
}
