class calico::policy_controller {

  include ::calico

  file { "${::calico::secure_config_dir}/calico-config.yaml":
    ensure  => file,
    content => template('calico/calico-config.yaml.erb'),
  } ->
  exec { 'deploy calico config':
    command => "${::calico::kubectl_bin} apply -f ${::calico::secure_config_dir}/calico-config.yaml",
    unless  => "${::calico::kubectl_bin} get -f ${::calico::secure_config_dir}/calico-config.yaml",
  } ->
  file { "${::calico::secure_config_dir}/policy-controller-deployment.yaml":
    ensure  => file,
    content => template('calico/policy-controller-deployment.yaml.erb'),
  } ->
  exec { 'deploy calico policy controller':
    command => "${::calico::kubectl_bin} apply -f ${::calico::secure_config_dir}/policy-controller-deployment.yaml",
    unless  => "${::calico::kubectl_bin} get -f ${::calico::secure_config_dir}/policy-controller-deployment.yaml",
  }
}
