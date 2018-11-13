class prometheus::kube_state_metrics (
  String $image = 'gcr.io/google_containers/kube-state-metrics',
  String $version = '1.4.0',
  String $resizer_image = 'gcr.io/google_containers/addon-resizer',
  String $resizer_version = '1.8.4',
){
  require ::kubernetes
  include ::prometheus

  $namespace = $::prometheus::namespace

  $authorization_mode = $::kubernetes::_authorization_mode
  if member($authorization_mode, 'RBAC'){
    $rbac_enabled = true
  } else {
    $rbac_enabled = false
  }

  if versioncmp($::kubernetes::version, '1.6.0') >= 0 {
    $version_before_1_6 = false
  } else {
    $version_before_1_6 = true
  }

  kubernetes::apply{'kube-state-metrics':
    manifests => [
      template('prometheus/prometheus-ns.yaml.erb'),
      template('prometheus/kube-state-metrics-deployment.yaml.erb'),
      template('prometheus/kube-state-metrics-service.yaml.erb'),
    ],
  }

  prometheus::rule { 'KubernetesPodUnready':
    expr        => '(kube_pod_info{created_by_kind!="Job"} and ON(pod, namespace) kube_pod_status_ready{condition="true"}) == 0',
    for         => '5m',
    summary     => '{{$labels.namespace}}/{{$labels.pod}}: pod is unready',
    description => '{{$labels.namespace}}/{{$labels.pod}}: pod is unready',
  }

  prometheus::rule { 'KubernetesPodFrequentlyRestarting':
    expr        => 'increase(kube_pod_container_status_restarts[1h]) > 5',
    for         => '5m',
    summary     => '{{$labels.namespace}}/{{$labels.pod}}: pod is too frequently restarting',
    description => 'Pod {{$labels.namespaces}}/{{$labels.pod}} was restarted {{$value}} times within the last hour',
  }

  prometheus::rule { 'KubernetesNodeUnready':
    expr        => 'SUM(kube_node_status_condition{status="true",condition="Ready"} * ON(node) group_right kube_node_labels) WITHOUT (kubernetes_name, kubernetes_namespace, job, app, instance, condition) == 0',
    for         => '5m',
    summary     => '{{$labels.node}}: node is unready',
    description => '{{$labels.node}}: node is unready {{$labels}}',
  }
}
