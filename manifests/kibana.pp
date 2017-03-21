class kubernetes_addons::kibana(
  String $namespace=$::kubernetes_addons::params::namespace,
  String $image='gcr.io/google_containers/kibana',
  String $version='v4.6.1-1',
  String $request_cpu='50m',
  String $request_mem='768Mi',
  String $limit_cpu='100m',
  String $limit_mem='2Gi',
  Integer $replicas=2,
) inherits ::kubernetes_addons::params {
  require ::kubernetes

  kubernetes::apply{'kibana':
    manifests => [
      template('kubernetes_addons/kibana-svc.yaml.erb'),
      template('kubernetes_addons/kibana-deployment.yaml.erb'),
    ],
  }
}
