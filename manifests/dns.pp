# == Class kubernetes::dns
class kubernetes::dns(
  $image='gcr.io/google_containers/k8s-dns-kube-dns-amd64',
  $version='1.14.2',
  $dnsmasq_image='gcr.io/google_containers/k8s-dns-dnsmasq-nanny-amd64',
  $dnsmasq_version='1.14.2',
  $sidecar_image='gcr.io/google_containers/k8s-dns-sidecar-amd64',
  $sidecar_version='1.14.2',
  $autoscaler_image='gcr.io/google_containers/cluster-proportional-autoscaler-amd64',
  $autoscaler_version='1.1.1-r2',
  $min_replicas=3,
){
  require ::kubernetes

  $cluster_domain = $::kubernetes::cluster_domain
  $cluster_dns = $::kubernetes::_cluster_dns

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

  kubernetes::apply{'kube-dns':
    manifests => [
      template('kubernetes/kube-dns-config-map.yaml.erb'),
      template('kubernetes/kube-dns-service-account.yaml.erb'),
      template('kubernetes/kube-dns-deployment.yaml.erb'),
      template('kubernetes/kube-dns-svc.yaml.erb'),
      template('kubernetes/kube-dns-horizontal-autoscaler-deployment.yaml.erb'),
      template('kubernetes/kube-dns-horizontal-autoscaler-rbac.yaml.erb'),
    ],
  }
}
