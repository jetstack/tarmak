class fluent_bit::config {

  if defined(Class['kubernetes::apiserver']) and $::kubernetes::apiserver::audit_enabled {
    $audit_enabled = true
    $audit_log_path = $::kubernetes::apiserver::audit_log_path
  } else {
    $audit_enabled = false
  }

  file { '/etc/td-agent-bit/td-agent-bit.conf':
    ensure  => file,
    mode    => '0644',
    owner   => 'root',
    group   => 'root',
    content => template('fluent_bit/td-agent-bit.conf.erb'),
  }

}
