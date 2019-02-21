define prometheus::scrape_config (
  Integer $order,
  $config = {},
  $job_name = $title,
  $ensure = $::prometheus::ensure,
) {
  include ::prometheus

  if ! defined(Class['kubernetes::apiserver']) {
    fail('This defined type can only be used on the kubernetes master')
  }

  kubernetes::apply_fragment { "prometheus-scrape-config-${job_name}":
    ensure  => $ensure,
    content => template('prometheus/prometheus-config-frag.yaml.erb'),
    order   => 400 + $order,
    target  => 'prometheus-config',
  }
}
