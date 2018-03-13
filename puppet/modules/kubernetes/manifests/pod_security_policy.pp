# This class manages RBAC manifests
class kubernetes::pod_security_policy{
  require ::kubernetes
  if $_pod_security_policy {
    kubernetes::apply{'pod-security-policy':
      manifests => [
        template('kubernetes/pod-security-policy.yaml.erb'),
      ],
    }
  }
}
