# This class manages RBAC manifests
class kubernetes::rbac{
  require ::kubernetes

  $authorization_mode = $kubernetes::_authorization_mode
  if member($authorization_mode, 'RBAC'){
    kubernetes::apply{'puppernetes-rbac-system:node':
      manifests => [
        template('kubernetes/rbac-crb-system:node.yaml.erb'),
      ],
    }
  }
}
