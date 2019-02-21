class kubernetes_addons::grafana(
  $image=$::kubernetes_addons::params::grafana_image,
  $version=$::kubernetes_addons::params::grafana_version,
  Enum['present', 'absent'] $ensure = 'present',
) inherits ::kubernetes_addons::params {
  require ::kubernetes

  kubernetes::apply{'heapster-grafana':
    ensure    => $ensure,
    manifests => [
      template('kubernetes_addons/grafana-svc.yaml.erb'),
      template('kubernetes_addons/grafana-deployment.yaml.erb'),
    ],
  }
}
