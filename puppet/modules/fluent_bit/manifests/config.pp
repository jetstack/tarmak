class fluent_bit::config {

  file { '/etc/td-agent-bit/td-agent-bit.conf':
    ensure  => file,
    mode    => '0644',
    owner   => 'root',
    group   => 'root',
    content => template('fluent_bit/td-agent-bit.conf.erb'),
  }

  $types = ['all']
  file { '/etc/td-agent-bit/td-agent-bit-output.conf':
    ensure  => file,
    mode    => '0644',
    owner   => 'root',
    group   => 'root',
    content => template('fluent_bit/td-agent-bit-output.conf.erb'),
  }

}
