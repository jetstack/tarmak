### Get blackbox_exporter
class prometheus::blackbox_etcd (
  $download_url = $::prometheus::blackbox_download_url,
  $dest_dir = $::prometheus::blackbox_dest_dir,
  $systemd_path = $::prometheus::systemd_path
)
{

#  $download_url = regsubst(
#    $::calico::params::calico_bin_download_url,
#    '#VERSION#',
#    $bin_version,
#    'G'
#    )

  file { $dest_dir:
    ensure => directory,
  } ->
  prometheus::wget_file { 'blackbox_exporter':
    url             => "${download_url}/blackbox_exporter",
    destination_dir => $dest_dir,
  } ->
  file { "${dest_dir}/blackbox_exporter":
    ensure => file,
    mode   => '0755',
  }

  file { "${dest_dir}/blackbox.yml":
    ensure  => file,
    content => template('prometheus/blackbox_exporter.yaml.erb'),
  }

  file { "${systemd_path}/blackbox_exporter.service":
    ensure  => file,
    content => template('prometheus/blackbox_exporter.service.erb'),
  } ~>
  exec { "${module_name}-systemctl-daemon-reload":
    command     => '/usr/bin/systemctl daemon-reload',
    refreshonly => true,
  }


  service { 'blackbox_exporter':
    ensure  => running,
    enable  => true,
    require => [ File[ "${dest_dir}/blackbox_exporter"], File["${dest_dir}/blackbox.yml"], File["${systemd_path}/blackbox_exporter.service"]],
  }
}
