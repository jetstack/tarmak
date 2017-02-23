class kubernetes_addons::nginx_ingress(
  $image=$::kubernetes_addons::params::nginx_ingress_image,
  $version=$::kubernetes_addons::params::nginx_ingress_version,
  $request_cpu=$::kubernetes_addons::params::nginx_ingress_request_cpu,
  $request_mem=$::kubernetes_addons::params::nginx_ingress_request_mem,
  $limit_cpu=$::kubernetes_addons::params::nginx_ingress_limit_cpu,
  $limit_mem=$::kubernetes_addons::params::nginx_ingress_limit_mem,
  $namespace=$::kubernetes_addons::params::namespace,
  $replicas=undef,
  $host_port=false,
) inherits ::kubernetes_addons::params {
  require ::kubernetes

  kubernetes::apply{'nginx-ingress':
    manifests => [
      template('kubernetes_addons/nginx-ingress-svc.yaml.erb'),
      template('kubernetes_addons/nginx-ingress-deployment.yaml.erb'),
    ],
  }
}
