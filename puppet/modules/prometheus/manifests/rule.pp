define prometheus::rule (
  String $expr,
  String $summary,
  String $description,
  String $for = '5m',
  $labels = {'severity' => 'page'},
  Integer $order = 10,
) {
  include ::prometheus

  if ! defined(Class['kubernetes::apiserver']) {
    fail('This defined type can only be used on the kubernetes master')
  }

  $config = {
    'groups' =>  [{
      'name'  => $name,
      'rules' =>  [
        {
          'alert'       => $name,
          'expr'        => $expr,
          'for'         => $for,
          'labels'      => $labels,
          'annotations' => {
            'summary'     => $summary,
            'description' => $description,
          }
        }
      ]
    }]
  }

  kubernetes::apply_fragment { "prometheus-rules-${title}":
    ensure  => $::prometheus::ensure,
    content => template('prometheus/prometheus-rule.yaml.erb'),
    order   => $order,
    target  => 'prometheus-rules',
  }
}
