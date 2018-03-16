# This class manages RBAC manifests
class kubernetes::pod_security_policy{
  require ::kubernetes
  
  $pod_security_policy = $::kubernetes::_pod_security_policy
  if $pod_security_policy {
    kubernetes::apply{'pod-security-policy':
      manifests => [
        template('kubernetes/pod-security-policy.yaml.erb'),
      ],
    }
  }
}
