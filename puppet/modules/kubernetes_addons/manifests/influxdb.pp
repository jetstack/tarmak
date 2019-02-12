class kubernetes_addons::influxdb(
  $image=$::kubernetes_addons::params::influxdb_image,
  $version=$::kubernetes_addons::params::influxdb_version,
  $enabled = true,
) inherits ::kubernetes_addons::params {
  require ::kubernetes

  if $enabled {
    kubernetes::apply{'heapster-influxdb':
      manifests => [
        template('kubernetes_addons/influxdb-svc.yaml.erb'),
        template('kubernetes_addons/influxdb-deployment.yaml.erb'),
      ],
    }
  } else {
    kubernetes::delete{'heapster-influxdb':}
  }
}
