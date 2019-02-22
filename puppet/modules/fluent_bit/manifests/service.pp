class fluent_bit::service {
  if $::fluent_bit::ensure == 'present' {
    $service_ensure = 'running'
    $service_enable = true
  } else {
    $service_ensure = 'stopped'
    $service_enable = false
  }

  service { $::fluent_bit::service_name:
    ensure     => $service_ensure,
    enable     => $service_enable,
    hasstatus  => true,
    hasrestart => true,
  }

}
