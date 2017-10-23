class kubernetes_addons::tiller(
  String $image='gcr.io/kubernetes-helm/tiller',
  String $version='v2.6.1',
) inherits ::kubernetes_addons::params {
  require ::kubernetes

  kubernetes::apply{'tiller':
    manifests => [
      template('kubernetes_addons/tiller-deployment.yaml.erb'),
    ],
  }
}
