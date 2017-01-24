class calico::policy_controller (
  $etcd_cert_path = $::calico::params::etcd_cert_path,
  $policy_controller_version = $::calico::params::policy_controller_version,
) inherits ::calico::params
{

  include ::calico

  file { "${::calico::params::helper_dir}/kubectl_helper_cali.sh":
    ensure  => file,
    content => template('calico/kubectl_helper.sh.erb'),
    mode    => '0755',
  } ->
  file { "${::calico::params::secure_config_dir}/calico-config.yaml":
    ensure  => file,
    content => template('calico/calico-config.yaml.erb'),
  } ->
  exec { 'deploy calico config':
    command => "${::calico::params::helper_dir}/kubectl_helper_cali.sh apply ${::calico::params::secure_config_dir}/calico-config.yaml",
    unless  => "${::calico::params::helper_dir}/kubectl_helper_cali.sh get ${::calico::params::secure_config_dir}/calico-config.yaml",
  } ->
  file { "${::calico::params::secure_config_dir}/policy-controller-deployment.yaml":
    ensure  => file,
    content => template('calico/policy-controller-deployment.yaml.erb'),
  } ->
  exec { 'deploy calico policy controller':
    command => "${::calico::params::helper_dir}/kubectl_helper_cali.sh apply ${::calico::params::secure_config_dir}/policy-controller-deployment.yaml",
    unless  => "${::calico::params::helper_dir}/kubectl_helper_cali.sh get ${::calico::params::secure_config_dir}/policy-controller-deployment.yaml",
  }
}
