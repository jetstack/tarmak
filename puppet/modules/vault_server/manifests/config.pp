# == Class vault_server::config
#
# This class is called from vault_server for service config.
#
class vault_server::config {
  if $::vault_server::init_token != undef {
    $init_token_enabled = true
  } else {
    $init_token_enabled = false
  }
  $server_url = $::vault_server::server_url
  $init_role = $::vault_server::init_role

  if $::vault_server::ca_cert_path != undef {
    $ca_cert_path = $::vault_server::ca_cert_path
  }

  file { $::vault_server::config_dir:
    ensure => directory,
    mode   => '0700',
  }

  ## if init token provided, get a unique token for node
  if $::vault_server::init_token != undef {
    file {$::vault_server::init_token_path:
      ensure  => 'present',
      replace => 'no',
      content => $::vault_server::init_token,
      mode    => '0600',
    }
  }

  ## if token provided, get a unique token for node
  if $::vault_server::token != undef {
    file {$::vault_server::token_path:
      ensure  => 'present',
      content => $::vault_server::token,
      mode    => '0600',
    }
  }

  file { $::vault_server::config_path:
    ensure  => file,
    content => template('vault_server/config.erb'),
  }

}
