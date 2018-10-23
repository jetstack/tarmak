# Create an instance of an etcd server
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
  String $advertise_client_network = nil,
  Array[String] $systemd_wants = [],
  Array[String] $systemd_requires = [],
  Array[String] $systemd_after = [],
  Array[String] $systemd_before = [],
  Array $initial_cluster = [],
  Optional[Boolean] $backup_enabled = undef,
  Optional[Enum['aws:kms','']]$backup_sse = undef,
  Optional[String] $backup_bucket_prefix = undef
) {
  include ::etcd

  $_systemd_wants = $systemd_wants
  $_systemd_after = $systemd_after
  $_systemd_requires = $systemd_after
  $_systemd_before = $systemd_before

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

  if $advertise_client_network != nil {
    $advertise_client_ip = get_ipaddress_in_network($advertise_client_network)
  } else {
    $advertise_client_ip = undef
  }

  if $advertise_client_ip != undef {
    $advertise_client_urls = "${proto}://${advertise_client_ip}:${client_port}"
  } else {
    $advertise_client_urls = undef
  }

  $listen_client_urls = "${proto}://0.0.0.0:${client_port}"

  if $members == 1 {
    $listen_peer_urls = "${proto}://127.0.0.1:${peer_port}"
  } else {
    $listen_peer_urls = "${proto}://0.0.0.0:${peer_port}"
    $initial_advertise_peer_urls = "${proto}://${nodename}:${peer_port}"
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

  exec { "${cluster_name}-systemctl-daemon-reload":
    command     => 'systemctl daemon-reload',
    refreshonly => true,
    path        => defined('$::path') ? {
      default => '/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/opt/bin',
      true    => $::path
    },
  }

  file { "${::etcd::systemd_dir}/${service_name}.service":
    ensure  => file,
    content => template('etcd/etcd.service.erb'),
    require => [
      Etcd::Install[$version],
      Class['etcd'],
    ],
    notify  => Exec["${cluster_name}-systemctl-daemon-reload"]
  }
  ~> service { "${service_name}.service":
    ensure  => running,
    enable  => true,
    require => [
      File[$data_dir],
      Exec["${cluster_name}-systemctl-daemon-reload"],
    ],
  }

  # instantiate backup if enabled
  if $backup_enabled == true or ($backup_enabled == undef and $::etcd::backup_enabled == true) {
    etcd::backup { $name:
      version       => $version,
      client_port   => $client_port,
      service_name  => $service_name,
      tls           => $tls,
      tls_cert_path => $tls_cert_path,
      tls_key_path  => $tls_key_path,
      tls_ca_path   => $tls_ca_path,
      sse           => $backup_sse,
      bucket_prefix => $backup_bucket_prefix,
    }
  }
}
