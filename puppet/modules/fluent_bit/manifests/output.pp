define fluent_bit::output (
  Hash $config = {},
){
  include ::fluent_bit

  $path = $::fluent_bit::path
  $types = $config['types']
  $elasticsearch = $config['elasticsearch']

  if $elasticsearch and $elasticsearch['tlsCA'] and $elasticsearch['tlsCA'] != '' {
    file { "/etc/td-agent-bit/ssl/${name}-ca.pem":
      ensure  => file,
      mode    => '0640',
      owner   => 'root',
      group   => 'root',
      content => $elasticsearch['tlsCA'],
    }
  }

  if $elasticsearch and $elasticsearch['amazonESProxy'] {
    ::aws_es_proxy::instance{ $name:
      tls          => $elasticsearch['tls'],
      dest_port    => $elasticsearch['port'],
      dest_address => $elasticsearch['host'],
      listen_port  => $elasticsearch['amazonESProxy']['port'],
    }
  } else {
    ::aws_es_proxy::instance{ $name:
      ensure_service => 'stopped',
      enable_service => false,
      dest_address   => '',
    }
  }

  file { "/etc/td-agent-bit/td-agent-bit-output-${name}.conf":
    ensure  => file,
    mode    => '0640',
    owner   => 'root',
    group   => 'root',
    content => template('fluent_bit/td-agent-bit-output.conf.erb'),
    notify  => Service[$::fluent_bit::service_name],
  }

  file { "/etc/td-agent-bit/daemonset/td-agent-bit-output-${name}.conf":
    ensure  => file,
    mode    => '0640',
    owner   => 'root',
    group   => 'root',
    content => template('fluent_bit/td-agent-bit-output.conf.erb'),
  }

}
