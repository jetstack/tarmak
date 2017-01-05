class calico::policy_controller (
  $etcd_cert_path = $::calico::etcd_cert_path,
  $policy_controller_version = $::calico::policy_controller_version,
) inherits ::calico
{

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
