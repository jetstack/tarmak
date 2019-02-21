# == Class kubernetes::dns
class kubernetes::dns(
  $image='gcr.io/google_containers/k8s-dns-kube-dns-amd64',
  $version='1.14.5',
  $dnsmasq_image='gcr.io/google_containers/k8s-dns-dnsmasq-nanny-amd64',
  $dnsmasq_version='1.14.5',
  $sidecar_image='gcr.io/google_containers/k8s-dns-sidecar-amd64',
  $sidecar_version='1.14.5',
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

  $post_1_10 = versioncmp($::kubernetes::version, '1.10.0') >= 0

  if $post_1_10 {
    $app_name = 'core-dns'
    $delete_app_name = 'kube-dns'
    $label_name = 'CoreDNS'

    $manifests = [
      template('kubernetes/core-dns-config-map.yaml.erb'),
      template('kubernetes/core-dns-deployment.yaml.erb'),
    ]

  } else {
    $app_name = 'kube-dns'
    $delete_app_name = 'core-dns'
    $label_name = 'KubeDNS'

    $manifests = [
      template('kubernetes/kube-dns-config-map.yaml.erb'),
      template('kubernetes/kube-dns-deployment.yaml.erb'),
    ]
  }

    kubernetes::apply{$app_name:
      ensure    => 'present',
      manifests => concat(
        $manifests,
        template('kubernetes/dns-service-account.yaml.erb'),
        template('kubernetes/dns-svc.yaml.erb'),
        template('kubernetes/dns-horizontal-autoscaler-deployment.yaml.erb'),
        template('kubernetes/dns-horizontal-autoscaler-rbac.yaml.erb'),
        template('kubernetes/dns-cluster-role.yaml.erb'),
        template('kubernetes/dns-cluster-role-binding.yaml.erb'),
      ),
    } -> kubernetes::apply{$delete_app_name:
      ensure    => 'absent',
      manifests => [],
    }
}
