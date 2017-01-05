class calico::lo_install (
  $cni_version = $::calico::calico_cni_version
) inherits ::calico
{

  include ::calico

  $dest_dir = "${::calico::install_dir}/bin"

  $download_url = regsubst(
    $::calico::params::cni_download_url,
    '#VERSION#',
    $cni_version,
    'G'
  )

  archive { 'download and extract cni-lo':
    source          => $download_url,
    path            => "/tmp/cni-${cni_version}.tgz",
    extract         => true,
    extract_path    => "${dest_dir}/",
    extract_command => 'tar -xzf %s ./loopback',
    creates         => "${dest_dir}/loopback",
    require         => Class['calico'],
  }
}
