class kubernetes_addons::tiller(
  $image=$::kubernetes_addons::params::tiller_image,
  $version=$::kubernetes_addons::params::tiller_version,
) inherits ::kubernetes_addons::params {
  require ::kubernetes

  kubernetes::apply{'tiller':
    manifests => [
      template('kubernetes_addons/tiller-deployment.yaml.erb'),
    ],
  }
}
