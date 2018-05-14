class fluent_bit::service {

  service { $::fluent_bit::service_name:
    ensure     => running,
    enable     => true,
    hasstatus  => true,
    hasrestart => true,
  }

}
