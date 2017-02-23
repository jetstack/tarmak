class kubernetes_addons::params{
  require ::kubernetes

  if defined('$kubernetes::cloud_provider') {
    $cloud_provider=$::kubernetes::cloud_provider
  } else {
    $cloud_provider=undef
  }

  if $cloud_provider == 'aws' {
    $aws_region = $::ec2_metadata['placement']['availability-zone'][0,-2]
  } else {
    $aws_region = undef
  }

  $ca_bundle_path='/etc/ssl/certs'
  $tiller_image = 'gcr.io/kubernetes-helm/tiller'
  $tiller_version = 'v2.2.0'

  $default_backend_image='gcr.io/google_containers/defaultbackend'
  $default_backend_version='1.2'
  $default_backend_limit_cpu='10m'
  $default_backend_limit_mem='20Mi'
  $default_backend_request_cpu='10m'
  $default_backend_request_mem='20Mi'

  $nginx_ingress_image='gcr.io/google_containers/nginx-ingress-controller'
  $nginx_ingress_version='0.8.3'
  $nginx_ingress_limit_cpu='200m'
  $nginx_ingress_limit_mem='300Mi'
  $nginx_ingress_request_cpu='100m'
  $nginx_ingress_request_mem='100Mi'

  $cluster_autoscaler_image='gcr.io/google_containers/cluster-autoscaler'
  $cluster_autoscaler_version='v0.4.0'
  $cluster_autoscaler_limit_cpu='200m'
  $cluster_autoscaler_limit_mem='500Mi'
  $cluster_autoscaler_request_cpu='100m'
  $cluster_autoscaler_request_mem='300Mi'
  $cluster_autoscaler_min_instances=3
  $cluster_autoscaler_max_instances=6
  if defined('$kubernetes::cluster_name') {
    $cluster_autoscaler_asg_name="kubernetes-${::kubernetes::cluster_name}-worker"
  } else {
    $cluster_autoscaler_asg_name=undef
  }

  $dashboard_image='gcr.io/google_containers/kubernetes-dashboard-amd64'
  $dashboard_version='v1.5.1'
  $dashboard_limit_cpu='100m'
  $dashboard_limit_mem='128Mi'
  $dashboard_request_cpu='10m'
  $dashboard_request_mem='64Mi'

  $heapster_image='gcr.io/google_containers/heapster-amd64'
  $heapster_version='v1.3.0-beta.1'
  $heapster_nanny_limit_cpu='50m'
  $heapster_nanny_limit_mem='100Mi'
  $heapster_nanny_request_cpu='50m'
  $heapster_nanny_request_mem='100Mi'
  $heapster_nanny_image='gcr.io/google_containers/addon-resizer'
  $heapster_nanny_version='1.7'
  $heapster_cpu='100m'
  $heapster_mem='150Mi'
  $heapster_extra_cpu='0.5m'
  $heapster_extra_mem='4Mi'

  $influxdb_image='luxas/heapster-influxdb-amd64'
  $influxdb_version='v0.13.0'

  $grafana_image='gcr.io/google_containers/heapster-grafana-amd64'
  $grafana_version='v4.0.2'

}
