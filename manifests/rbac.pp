# This class manages RBAC manifests
class kubernetes::rbac{
  require ::kubernetes

  $authorization_mode = $kubernetes::_authorization_mode
  if member($authorization_mode, 'RBAC'){
    kubernetes::apply{'puppernetes-rbac':
      manifests => [
        template('kubernetes/puppernetes-rbac.yaml.erb'),
      ],
    }
  }
}
