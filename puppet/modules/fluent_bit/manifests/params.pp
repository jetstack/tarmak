class fluent_bit::params(
  String $version = '1.0.4',
){
  $package_name = 'td-agent-bit'
  $service_name = 'td-agent-bit'
  # After updating this version you need to make sure you run the scripts in
  # /hack/fluentbit-repo/ to clone their repo and lock the version
}
