define etcd::instance (
  String $version,
  Integer $client_port = 2379,
  Integer $peer_port = 2380,
  Integer $members = 1,
  String $nodename= $::fqdn,
  Boolean $tls = false,
  String $tls_cert_path = nil,
  String $tls_key_path = nil,
  String $tls_ca_path = nil,
  Array $initial_cluster = []
){
  include ::etcd
  include ::systemd

  $user = $::etcd::user
  $group = $::etcd::group
  $cluster_name = $name
  $service_name = "etcd-${cluster_name}"
  $data_dir = "${::etcd::data_dir}/${cluster_name}"

  if $tls {
    $proto = 'https'
  } else {
    $proto = 'http'
  }

  $advertise_client_urls = "${proto}://0.0.0.0:${client_port}"
  $listen_client_urls = "${proto}://0.0.0.0:${client_port}"

  if $members == 1 {
    $listen_peer_urls = "${proto}://127.0.0.1:${peer_port}"
  } else {
    $listen_peer_urls = "${proto}://0.0.0.0:${peer_port}"
    $initial_advertise_peer_urls = "${proto}://${::fqdn}:${peer_port}"
    $_initial_cluster = $initial_cluster.map |$node| { "${node}=${proto}://${node}:${peer_port}" }.join(',')
    $_initial_cluster_hash = md5($_initial_cluster)
    $initial_cluster_token = "etcd-${cluster_name}-${_initial_cluster_hash}"
    $initial_cluster_state = 'new'
  }

  ensure_resource('etcd::install', $version, {
    ensure  => present,
    require => Class['etcd'],
  })

  file { $data_dir:
    ensure => directory,
    owner  => $::etcd::user,
    group  => $::etcd::group,
    mode   => '0750',
  }

  file { "/etc/systemd/system/${service_name}.service":
    ensure  => file,
    content => template('etcd/etcd.service.erb'),
    require => [
      Etcd::Install[$version],
      Class['etcd'],
    ],
    notify  => Exec['systemctl-daemon-reload']
  } ~>
  service { "${service_name}.service":
    ensure  => running,
    enable  => true,
    require => [
      File[$data_dir],
      Exec['systemctl-daemon-reload'],
    ],
  }

}
