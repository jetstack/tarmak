class calico::bin_install {

  include ::calico

  $version = $::calico::params::calico_bin_version

  $dest_dir = "${::calico::install_dir}/bin"

  $download_url = regsubst(
    $::calico::params::calico_bin_download_url,
    '#VERSION#',
    $version,
    'G'
    )

  notify{"Calico bin download url is ${download_url}/calico":}

  wget::fetch { "calico-v${version}":
    source      => "${download_url}/calico",
    destination => $dest_dir,
    require     => Class['calico'],
    before      => File["${dest_dir}/calico"],
  }

  wget::fetch { "calico-ipam-v${version}":
    source      => "${download_url}/calico-ipam",
    destination => $dest_dir,
    require     => Class['calico'],
    before      => File["${dest_dir}/calico-ipam"],
  }

  file { ["${dest_dir}/calico","${dest_dir}/calico-ipam"]:
    ensure => file,
    mode   => '0755',
  }
}
