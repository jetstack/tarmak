# == Class kubernetes::dns
class kubernetes::dns(
  $image='gcr.io/google_containers/kubedns-amd64',
  $version='1.9',
  $dnsmasq_image='gcr.io/google_containers/kube-dnsmasq-amd64',
  $dnsmasq_version='1.4',
  $dnsmasq_metrics_image='gcr.io/google_containers/dnsmasq-metrics-amd64',
  $dnsmasq_metrics_version='1.0',
  $exechealthz_image='gcr.io/google_containers/exechealthz-amd64',
  $exechealthz_version='1.2',
  $autoscaler_image='gcr.io/google_containers/cluster-proportional-autoscaler-amd64',
  $autoscaler_version='1.0.0',
  $min_replicas=3,
){
  require ::kubernetes

  $cluster_domain = $::kubernetes::cluster_domain
  $cluster_dns = $::kubernetes::_cluster_dns

  kubernetes::apply{'kube-dns':
    manifests => [
      template('kubernetes/kube-dns-deployment.yaml.erb'),
      template('kubernetes/kube-dns-svc.yaml.erb'),
      template('kubernetes/kube-dns-horizontal-autoscaler-deployment.yaml.erb'),
    ],
  }
}
