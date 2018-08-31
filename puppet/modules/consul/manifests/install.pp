class consul::install(

){

  $_download_url = regsubst(
    $consul::download_url,
    '#VERSION#',
    $consul::version,
    'G'
  )

  $_sha256sums_url = regsubst(
    $consul::sha256sums_url,
    '#VERSION#',
    $consul::version,
    'G'
  )

  $_signature_url = regsubst(
    $consul::signature_url,
    '#VERSION#',
    $consul::version,
    'G'
  )

  # install unzip if necessary
  ensure_resource('package', 'unzip', {})

  $zip_path = "${consul::_dest_dir}/consul.zip"

  Package['unzip']
  -> Archive[$zip_path]

  file {$consul::_dest_dir:
    ensure => directory,
    mode   => '0755',
  }
  -> archive {$zip_path:
    ensure           => present,
    extract          => true,
    source           => $_download_url,
    signature_binary => $_signature_url,
    sha256sums       => $_sha256sums_url,
    provider         => 'airworthy',
    extract_path     => $consul::_dest_dir,
  }
  -> file {$consul::bin_path:
    ensure => file,
    mode   => '0755',
  }
  -> file {"${consul::link_path}/${consul::app_name}":
    ensure => link,
    target => $consul::bin_path,
  }

  # install consul exporter if enabled
  if $consul::exporter_enabled {
    $exporter_tar_path = "${consul::exporter_dest_dir}/consul-exporter.tar.gz"

    $_exporter_download_url = regsubst(
      $consul::exporter_download_url,
      '#VERSION#',
      $consul::exporter_version,
      'G'
    )

    $_exporter_signature_url = regsubst(
      $consul::exporter_signature_url,
      '#VERSION#',
      $consul::exporter_version,
      'G'
    )

    file {$consul::exporter_dest_dir:
      ensure => directory,
      mode   => '0755',
    }
    -> archive {$exporter_tar_path:
      ensure            => present,
      extract           => true,
      source            => $_exporter_download_url,
      signature_armored => $_exporter_signature_url,
      provider          => 'airworthy',
      extract_path      => $consul::exporter_dest_dir,
      extract_command   => 'tar xfz %s --strip-components=1'
    }
  }
  else {
    file {$consul::exporter_dest_dir:
      ensure => absent,
    }
  }

  file { "${consul::_dest_dir}/consul-backup.sh":
    ensure  => file,
    content => file('consul/consul-backup.sh'),
    mode    => '0755'
  }
  -> file {"${consul::link_path}/consul-backup.sh":
    ensure => link,
    target => "${consul::_dest_dir}/consul-backup.sh",
  }
}
