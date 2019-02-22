# This class manages RBAC manifests
class kubernetes::pod_security_policy{
  require ::kubernetes

  $pod_security_policy = $::kubernetes::_pod_security_policy
  if $pod_security_policy {
    $ensure = 'present'

    $authorization_mode = $kubernetes::_authorization_mode
    if ! member($authorization_mode, 'RBAC') {
      fail('RBAC should be enabled when PodSecurityPolicy is enabled.')
    }

  } else {
    $ensure = 'absent'
  }

  kubernetes::apply{'puppernetes-rbac-psp':
    ensure    => $ensure,
    manifests => [
      template('kubernetes/pod-security-policy-rbac.yaml.erb'),
      template('kubernetes/pod-security-policy.yaml.erb'),
    ],
  }
}
