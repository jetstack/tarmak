class kubernetes_addons::fluentd_elasticsearch(
  String $namespace=$::kubernetes_addons::params::namespace,
  String $image='gcr.io/google_containers/fluentd-elasticsearch',
  String $version='1.22',
  String $request_cpu='200m',
  String $request_mem='384Mi',
  String $limit_cpu='100m',
  String $limit_mem='256Mi',
) inherits ::kubernetes_addons::params {
  require ::kubernetes

  kubernetes::apply{'fluentd-elasticsearch':
    manifests => [
      template('kubernetes_addons/fluentd-elasticsearch-daemonset.yaml.erb'),
    ],
  }
}
