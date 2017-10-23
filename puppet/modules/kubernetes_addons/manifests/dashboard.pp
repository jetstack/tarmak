class kubernetes_addons::dashboard(
  String $image='gcr.io/google_containers/kubernetes-dashboard-amd64',
  String $version='v1.5.1',
  String $limit_cpu='100m',
  String $limit_mem='128Mi',
  String $request_cpu='10m',
  String $request_mem='64Mi',
  $replicas=undef,
) inherits ::kubernetes_addons::params {
  require ::kubernetes

  kubernetes::apply{'kube-dashboard':
    manifests => [
      template('kubernetes_addons/dashboard-svc.yaml.erb'),
      template('kubernetes_addons/dashboard-deployment.yaml.erb'),
    ],
  }
}
