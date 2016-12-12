define etcd::install (
  String $ensure = 'present',
)
{
  include ::etcd

  $version = $name

  $dest_dir = "${::etcd::dest_dir}/${::etcd::params::app_name}-${version}"

  $download_url = regsubst(
    $::etcd::params::download_url,
    '#VERSION#',
    $version,
    'G'
  )

  $tar_path ="${dest_dir}/etcd-${version}.tar.gz"

  $etcd_bin = "${dest_dir}/etcd"

  file {$dest_dir:
    ensure => directory,
    mode   => '0755',
  } ->
  exec {"etcd-${version}-download":
    command => "curl -sL -o ${tar_path} ${download_url}",
    creates => $tar_path,
    path    => ['/usr/bin/', '/bin'],
  } ->
  exec {"etcd-${version}-extract":
    command => "tar xzf ${tar_path} --strip-components=1 -C ${dest_dir} --no-same-owner",
    creates => "${dest_dir}/etcd",
    path    => ['/usr/bin/', '/bin'],
  }



  #file { $dest_dir:
  #  ensure => directory,
  #  mode   => '0755',
  #} ->
  #archive { "${::etcd::download_dir}/etcd-${version}.tar.gz":
  ##  ensure       => present,
  #  extract      => true,
  #  extract_path => $dest_dir,
  #  cleanup      => true,
  #  creates      => $etcd_bin,
  #  require      => Class['etcd']
  #}
}
