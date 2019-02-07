class fluent_bit::daemonset(
  String $fluent_bit_image = 'fluent/fluent-bit',
  Array[String] $platform_namespaces = ['kube-system','service-broker','monitoring'],
){
  include ::fluent_bit::params
  require ::kubernetes

  $fluent_bit_version = $::fluent_bit::params::version
  $fluent_bit_image_tag = $::fluent_bit::params::version
  $busybox_image_tag = '1.30.0'

  $namespace = 'kube-system'

  $_sir_parts = $::kubernetes::service_ip_range_network.split('\.')
  $kubernetes_service_ip = "${_sir_parts[0]}.${_sir_parts[1]}.${_sir_parts[2]}.1"

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
