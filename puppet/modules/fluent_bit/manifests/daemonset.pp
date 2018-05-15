class fluent_bit::daemonset(
  String $fluent_bit_image = 'fluent/fluent-bit',
  String $fluent_bit_version = '0.13.1',
  Array[String] $platform_namespaces = ['kube-system','service-broker','monitoring'],
){
  require ::kubernetes

  $namespace = 'kube-system'

  $authorization_mode = $::kubernetes::_authorization_mode
  if member($authorization_mode, 'RBAC'){
    $rbac_enabled = true
  } else {
    $rbac_enabled = false
  }

  if versioncmp($::kubernetes::version, '1.6.0') >= 0 {
    $version_before_1_6 = false
  } else {
    $version_before_1_6 = true
  }

  kubernetes::apply{'fluent-bit':
    manifests => [
      template('fluent_bit/fluent-bit-configmap.yaml.erb'),
      template('fluent_bit/fluent-bit-daemonset.yaml.erb'),
    ]
  }
}
