# == Class vault_client::config
#
# This class is called from vault_client for service config.
#
class vault_client::config {
  if $::vault_client::init_token != undef {
    $init_token_enabled = true
  } else {
    $init_token_enabled = false
  }
  $server_url = $::vault_client::server_url
  $init_role = $::vault_client::init_role

  if $::vault_client::ca_cert_path != undef {
    $ca_cert_path = $::vault_client::ca_cert_path
  }

  file { $::vault_client::config_dir:
    ensure => directory,
    mode   => '0700',
  }

  ## if init token provided, get a unique token for node
  if $::vault_client::init_token != undef {
    file {$::vault_client::init_token_path:
      ensure  => 'present',
      replace => 'no',
      content => $::vault_client::init_token,
      mode    => '0600',
    }
  }

  ## if token provided, get a unique token for node
  if $::vault_client::token != undef {
    file {$::vault_client::token_path:
      ensure  => 'present',
      content => $::vault_client::token,
      mode    => '0600',
    }
  }

  file { $::vault_client::config_path:
    ensure  => file,
    content => template('vault_client/config.erb'),
  }
}
