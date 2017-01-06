class calico::lo_install (
  $cni_version = $::calico::params::calico_cni_version
) inherits ::calico::params
{
  $dest_dir = "${::calico::params::install_dir}/bin"

  $download_url = regsubst(
    $::calico::params::cni_download_url,
    '#VERSION#',
    $cni_version,
    'G'
  )

  archive { 'download and extract cni-lo':
    source          => $download_url,
    path            => "${::calico::params::tmp_dir}/cni-${cni_version}.tgz",
    extract         => true,
    extract_path    => "${dest_dir}/",
    extract_command => 'tar -xzf %s ./loopback',
    creates         => "${dest_dir}/loopback",
    require         => Class['calico'],
  }
}
