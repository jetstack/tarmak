# This class manages RBAC manifests
class kubernetes::pod_security_policy{
  require ::kubernetes
  
  $pod_security_policy = $::kubernetes::_pod_security_policy
  if $pod_security_policy {

    if ! member($authorization_mode, 'RBAC') {
      fail('RBAC should be enabled when PodSecurityPolicy is enabled.')
    }

    kubernetes::apply{'puppernetes-rbac-psp':
      manifests => [
        template('kubernetes/pod-security-policy-rbac.yaml.erb'),
        template('kubernetes/pod-security-policy.yaml.erb'),
      ],
    }
  }
}
