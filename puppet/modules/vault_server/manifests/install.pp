class vault_server::install(
  String $_download_url = $vault_server::params::download_url,
  String $_sha256sums_url = $vault_server::params::sha256sums_url,
  String $_signature_url = $vault_server::params::signature_url,
  String $_unsealer_download_url = $vault_server::params::unsealer_download_url,
  String $unsealer_sha256 = $vault_server::params::unsealer_sha256,
  String $consul_master_token = $vault_server::_consul_master_token,
  String $environment = $vault_server::environment,
  String $user = 'root',
  String $group = 'root',
)
{

  ## build download URL
  $download_url = regsubst(
    $_download_url,
    '#VERSION#',
    $vault_server::version,
    'G'
  )

  $sha256sums_url = regsubst(
    $_sha256sums_url,
    '#VERSION#',
    $vault_server::version,
    'G'
  )

  $signature_url = regsubst(
    $_signature_url,
    '#VERSION#',
    $vault_server::version,
    'G'
  )

  ## build download URL
  $unsealer_download_url = regsubst(
    $_unsealer_download_url,
    '#VERSION#',
    $vault_server::unsealer_version,
    'G'
  )

  # install unzip if necessary
  ensure_resource('package', 'unzip', {})

  $zip_path = "${vault_server::_dest_dir}/vault.zip"
  $vault_script_name = 'vault'

  Package['unzip']
  -> Archive[$zip_path]

  file {$vault_server::_dest_dir:
    ensure => directory,
    mode   => '0755',
  }
  -> archive {$zip_path:
    ensure           => present,
    extract          => true,
    source           => $download_url,
    signature_binary => $signature_url,
    sha256sums       => $sha256sums_url,
    provider         => 'airworthy',
    extract_path     => $vault_server::_dest_dir,
  }
  -> file {$vault_server::bin_path:
    mode   => '0755',
  }
  -> file {"${vault_server::link_path}/${vault_server::app_name}":
    ensure => link,
    target => $vault_server::bin_path,
  }

  archive {"${vault_server::_dest_dir}/vault-unsealer":
    ensure        => present,
    extract       => false,
    source        => $unsealer_download_url,
    checksum      => $unsealer_sha256,
    checksum_type => 'sha256',
    cleanup       => false,
    creates       => "${vault_server::_dest_dir}/vault-unsealer",
  }
  -> file {$vault_server::unsealer_bin_path:
    mode   => '0755',
  }
  -> file {"${vault_server::link_path}/${vault_server::app_name}-unsealer":
    ensure => link,
    target => $vault_server::unsealer_bin_path,
  }

  file { "/etc/profile.d/${vault_script_name}.sh":
    ensure  => file,
    content => file('vault_server/vault.sh'),
    owner   => $user,
    group   => $group,
    mode    => '0644',
  }

  file { "${vault_server::config_dir}/vault.hcl":
    ensure  => file,
    content => template('vault_server/vault.hcl.erb'),
    mode    => '0600'
  }

  file { "${vault_server::config_dir}/tls":
    ensure => directory,
    mode   => '0700'
  }
}
