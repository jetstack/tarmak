class prometheus::node_exporter_service (
  $systemd_path = $::prometheus::systemd_path,
  $node_exporter_image = $::prometheus::node_exporter_image,
  $node_exporter_version = $::prometheus::node_exporter_version,
  $node_exporter_port = $::prometheus::node_exporter_port
)
{
  include ::systemd

  file { "${systemd_path}/prometheus-node-exporter.service":
    ensure  => file,
    content => template('prometheus/prometheus-node-exporter.service.erb'),
    notify  => Exec['systemd-daemon-reload'],
  } ->
  service { 'prometheus-node-exporter':
    ensure => running,
    enable => true,
  }
}
