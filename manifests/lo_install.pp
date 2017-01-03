class calico::lo_install
{

  include ::calico

  $version = $::calico::params::calico_cni_version

  $dest_dir = "${::calico::install_dir}/bin"

  $download_url = regsubst(
    $::calico::params::cni_download_url,
    '#VERSION#',
    $version,
    'G'
  )

  archive { "download and extract cni-lo version ${version}":
    source          => $download_url,
    path            => "/tmp/cni-v${version}.tgz",
    extract         => true,
    extract_path    => "${dest_dir}/",
    extract_command => 'tar -xzf %s ./loopback',
    creates         => "${dest_dir}/loopback",
    require         => Class['calico'],
  }
}
