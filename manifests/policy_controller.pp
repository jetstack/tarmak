class calico::policy_controller (
  $etcd_cert_path = $::calico::params::etcd_cert_path,
  $policy_controller_version = $::calico::params::policy_controller_version,
) inherits ::calico::params
{
  include ::kubernetes
  include ::calico

  kubernetes::apply{'calico-policy-controller':
    manifests => [
      template('calico/calico-config.yaml.erb'),
      template('calico/policy-controller-deployment.yaml.erb'),
    ],
  }
}
