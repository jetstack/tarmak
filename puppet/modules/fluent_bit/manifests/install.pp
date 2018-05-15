class fluent_bit::install {

  include stdlib

  ensure_resource('package', [$::fluent_bit::package_name, 'curl'],{
    ensure => present
  })

  case $::osfamily {
    'RedHat': {
      file { '/etc/yum.repos.d/td-agent-bit.repo':
        ensure  => file,
        mode    => '0644',
        owner   => 'root',
        group   => 'root',
        require => Package['curl'],
        content => file('fluent_bit/td-agent-bit.repo'),
      }
    }
    'Debian': {
      exec { 'fluent_bit-add-gpg-key':
        command => 'curl -sL - https://packages.fluentbit.io/fluentbit.key | apt-key add -',
        require => Package['curl'],
        path    => $::fluent_bit::path,
      }
      # TODO: this doesn't look right
      -> exec { 'fluent_bit-add-source':
        command => 'deb http://packages.fluentbit.io/ubuntu xenial main',
        path    => $::fluent_bit::path,
      }
    }
    default: {
      fail("unsupported osfamily ${::osfamily}" )
    }
  }
  -> Package[$::fluent_bit::package_name]
  -> file { '/etc/td-agent-bit/daemonset':
    ensure => directory,
    mode   => '0755',
  }

}
