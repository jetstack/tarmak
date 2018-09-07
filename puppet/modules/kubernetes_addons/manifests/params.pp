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

  if $::osfamily == 'RedHat' {
    $ca_mounts=[
      {
        name      => 'ssl-certs',
        readOnly  => true,
        mountPath => '/etc/ssl/certs',
      },
      {
        name      => 'ssl-pki',
        readOnly  => true,
        mountPath => '/etc/pki',
      },
    ]
  }
  else {
    $ca_mounts=[
      {
        name      => 'ssl-certs',
        readOnly  => true,
        mountPath => '/etc/ssl/certs',
      },
    ]
  }

  $namespace = 'kube-system'

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

  $heapster_image='gcr.io/google_containers/heapster-amd64'
  $heapster_version='v1.5.4'
  $heapster_nanny_limit_cpu='50m'
  $heapster_nanny_limit_mem='100Mi'
  $heapster_nanny_request_cpu='50m'
  $heapster_nanny_request_mem='100Mi'
  $heapster_nanny_image='gcr.io/google_containers/addon-resizer'
  $heapster_nanny_version='1.8.1'
  $heapster_cpu='100m'
  $heapster_mem='150Mi'
  $heapster_extra_cpu='0.5m'
  $heapster_extra_mem='4Mi'

  $influxdb_image='gcr.io/google_containers/heapster-influxdb-amd64'
  $influxdb_version='v1.1.1'

  $grafana_image='gcr.io/google_containers/heapster-grafana-amd64'
  $grafana_version='v4.0.2'

}
