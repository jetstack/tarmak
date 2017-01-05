class calico::bin_install (
  $bin_version = $::calico::calico_bin_version
) inherits ::calico
{

  include ::calico

  $dest_dir = "${::calico::install_dir}/bin"

  $download_url = regsubst(
    $::calico::params::calico_bin_download_url,
    '#VERSION#',
    $bin_version,
    'G'
    )

  calico::wget_file { 'calico':
    url             => "${download_url}/calico",
    destination_dir => $dest_dir,
    require         => Class['calico'],
    before          => File["${dest_dir}/calico"],
  }

  calico::wget_file { 'calico-ipam':
    url             => "${download_url}/calico-ipam",
    destination_dir => $dest_dir,
    require         => Class['calico'],
    before          => File["${dest_dir}/calico-ipam"],
  }

  file { ["${dest_dir}/calico","${dest_dir}/calico-ipam"]:
    ensure => file,
    mode   => '0755',
  }
}
