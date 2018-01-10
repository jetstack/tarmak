define prometheus::scrape_config (
  Integer $order,
  $config = {},
  $job_name = $title,
) {
  if ! defined(Class['kubernetes::apiserver']) {
    fail('This defined type can only be used on the kubernetes master')
  }

  kubernetes::apply_fragment { "prometheus-scrape-config-${job_name}":
    content => template('prometheus/prometheus-config-frag.yaml.erb'),
    order   => 400 + $order,
    target  => 'prometheus-config',
  }
}
