class kubernetes_addons::elasticsearch(
  String $namespace=$::kubernetes_addons::params::namespace,
  String $image='gcr.io/google_containers/elasticsearch',
  String $version='v2.4.1-1',
  Boolean $persistent_storage=false,
  String $persistent_storage_request= '20Gi',
  String $persistent_storage_class= 'fast',
  String $request_cpu='100m',
  String $request_mem='512Mi',
  String $limit_cpu='1000m',
  String $limit_mem='2048Mi',
  Integer[0,65535] $node_port=0,
  Integer $replicas=2,
) inherits ::kubernetes_addons::params {
  require ::kubernetes

  # TODO: Support elasticsearch using StatefulSet pods
  kubernetes::apply{'elasticsearch':
    manifests => [
      template('kubernetes_addons/elasticsearch-svc.yaml.erb'),
      template('kubernetes_addons/elasticsearch-deployment.yaml.erb'),
    ],
  }
}
