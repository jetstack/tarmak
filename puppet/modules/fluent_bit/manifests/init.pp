class fluent_bit (
  $package_name = $::fluent_bit::params::package_name,
  $service_name = $::fluent_bit::params::service_name,
) inherits ::fluent_bit::params {

  $path = defined('$::path') ? {
    default => '/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/opt/bin',
    true    => $::path,
  }

  class { '::fluent_bit::install': }
  -> class { '::fluent_bit::config': }
  ~> class { '::fluent_bit::service': }
  -> Class['::fluent_bit']

}
