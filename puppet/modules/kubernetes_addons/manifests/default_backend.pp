class kubernetes_addons::default_backend(
  $image=$::kubernetes_addons::params::default_backend_image,
  $version=$::kubernetes_addons::params::default_backend_version,
  $request_cpu=$::kubernetes_addons::params::default_backend_request_cpu,
  $request_mem=$::kubernetes_addons::params::default_backend_request_mem,
  $limit_cpu=$::kubernetes_addons::params::default_backend_limit_cpu,
  $limit_mem=$::kubernetes_addons::params::default_backend_limit_mem,
  $namespace=$::kubernetes_addons::params::namespace,
  $replicas=undef,
) inherits ::kubernetes_addons::params {
  require ::kubernetes

  kubernetes::apply{'default-backend':
    manifests => [
      template('kubernetes_addons/default-backend-svc.yaml.erb'),
      template('kubernetes_addons/default-backend-deployment.yaml.erb'),
    ],
  }
}
