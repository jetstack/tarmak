# Sets up a blackbox exporter to blackbox probe in-cluster services and pods
class prometheus::blackbox_exporter(
  String $image = 'prom/blackbox-exporter',
  String $version = '0.12.0',
  Integer $port = 9115,
  Integer $replicas = 2,
)
{
  include ::prometheus
  $namespace = $::prometheus::namespace

  # Setup deployment for blackbox exporter in cluster
  if $::prometheus::role == 'master' {
    kubernetes::apply{'blackbox-exporter':
      manifests => [
        template('prometheus/prometheus-ns.yaml.erb'),
        template('prometheus/blackbox-exporter-deployment.yaml.erb'),
        template('prometheus/blackbox-exporter-svc.yaml.erb'),
      ],
    }
  }
}
