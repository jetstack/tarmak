# download and install airworthy
class airworthy::install {
  include airworthy

  $airworthy_path = "${::airworthy::_dest_dir}/airworthy"

  file { $::airworthy::_dest_dir:
    ensure => directory,
    mode   => '0755',
  }
  -> archive { $airworthy_path:
    ensure        => present,
    extract       => false,
    source        => $::airworthy::_download_url,
    checksum      => $::airworthy::_checksum,
    checksum_type => $::airworthy::checksum_type,
    creates       => $airworthy_path,
    cleanup       => false,
  }
  -> file {$airworthy_path:
    ensure => file,
    mode   => '0755',
    owner  => 'root',
    group  => 'root',
  }

  ensure_resource('file', [$::airworthy::bin_dir], {
    ensure => directory,
    mode   => '0755',
  })

  file { "${::airworthy::bin_dir}/airworthy":
    ensure  => link,
    target  => $airworthy_path,
    require => File[$::airworthy::bin_dir],
  }

  file { '/bin/airworthy':
    ensure  => link,
    target  => $airworthy_path,
    require => File[$::airworthy::bin_dir],
  }
}
