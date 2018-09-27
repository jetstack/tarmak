class tarmak::etcd(
){
  include ::tarmak
  include ::vault_client

  if $::tarmak::cloud_provider == 'aws' {
    $disks = aws_ebs::disks()
    case $disks.length {
      0: {$ebs_device = ''; $is_not_attached = true}
      1: {$ebs_device = 'xvdd'; $is_not_attached = true}
      default: {$ebs_device = $disks[1]; $is_not_attached = false}
    }

    class{'::aws_ebs':
      bin_dir     => $::tarmak::bin_dir,
      systemd_dir => $::tarmak::systemd_dir,
    }
    aws_ebs::mount{'etcd-data':
      volume_id     => $::tarmak_volume_id,
      device        => "/dev/${ebs_device}",
      dest_path     => '/var/lib/etcd',
      is_not_attached => $is_not_attached,
    } -> Etcd::Instance  <||>
  }

  file { $::tarmak::etcd_ssl_dir:
    ensure  => directory,
    owner   => $::tarmak::etcd_user,
    group   => $::tarmak::etcd_group,
    mode    => '0750',
    require => [ Class['etcd'] ],
  }

  $nodename = "${::tarmak::hostname}.${::tarmak::cluster_name}.${::tarmak::dns_root}"
  $alt_names = unique([
    $nodename,
    $::fqdn,
    'localhost',
  ])
  $ip_sans = unique([
    '127.0.0.1',
    $::ipaddress,
  ])

  if ! ($nodename in $::tarmak::_etcd_cluster) {
    fail("The node ${nodename} is not within the etcd_cluster (${tarmak::_etcd_cluster})")
  }

  vault_client::cert_service { 'etcd-k8s-main':
    base_path   => "${::tarmak::etcd_ssl_dir}/${::tarmak::etcd_k8s_main_ca_name}",
    common_name => 'etcd-server',
    alt_names   => $alt_names,
    ip_sans     => $ip_sans,
    role        => "${tarmak::cluster_name}/pki/${::tarmak::etcd_k8s_main_ca_name}/sign/server",
    uid         => $::tarmak::etcd_uid,
    exec_post   => [
      "-${::tarmak::systemctl_path} --no-block try-restart etcd-k8s-main.service etcd-k8s-events.service"
    ],
  }

  vault_client::cert_service { 'etcd-k8s-overlay':
    base_path   => "${::tarmak::etcd_ssl_dir}/${::tarmak::etcd_overlay_ca_name}",
    common_name => 'etcd-server',
    alt_names   => $alt_names,
    ip_sans     => $ip_sans,
    role        => "${tarmak::cluster_name}/pki/${::tarmak::etcd_overlay_ca_name}/sign/server",
    uid         => $::tarmak::etcd_uid,
    require     => [ User[$::tarmak::etcd_user], File[$::tarmak::etcd_ssl_dir] ],
    exec_post   => [
      "-${::tarmak::systemctl_path} --no-block try-restart etcd-overlay.service"
    ],
  }

  Class['vault_client'] -> Class['tarmak::etcd']

  class{'etcd':
    user  => $::tarmak::etcd_user,
    group => $::tarmak::etcd_group
  }

  etcd::instance{'k8s-main':
    version                  => $::tarmak::etcd_k8s_main_version,
    nodename                 => $nodename,
    members                  => $::tarmak::etcd_instances,
    initial_cluster          => $::tarmak::_etcd_cluster,
    advertise_client_network => $::tarmak::etcd_advertise_client_network,
    client_port              => $::tarmak::etcd_k8s_main_client_port,
    peer_port                => $::tarmak::etcd_k8s_main_peer_port,
    tls                      => true,
    tls_cert_path            => "${::tarmak::etcd_ssl_dir}/${::tarmak::etcd_k8s_main_ca_name}.pem",
    tls_key_path             => "${::tarmak::etcd_ssl_dir}/${::tarmak::etcd_k8s_main_ca_name}-key.pem",
    tls_ca_path              => "${::tarmak::etcd_ssl_dir}/${::tarmak::etcd_k8s_main_ca_name}-ca.pem",
    systemd_after            => delete_undef_values([$::tarmak::etcd_mount_unit]),
    systemd_requires         => delete_undef_values([$::tarmak::etcd_mount_unit]),
  }
  etcd::instance{'k8s-events':
    version                  => $::tarmak::etcd_k8s_events_version,
    nodename                 => $nodename,
    members                  => $::tarmak::etcd_instances,
    initial_cluster          => $::tarmak::_etcd_cluster,
    advertise_client_network => $::tarmak::etcd_advertise_client_network,
    client_port              => $::tarmak::etcd_k8s_events_client_port,
    peer_port                => $::tarmak::etcd_k8s_events_peer_port,
    tls                      => true,
    tls_cert_path            => "${::tarmak::etcd_ssl_dir}/${::tarmak::etcd_k8s_events_ca_name}.pem",
    tls_key_path             => "${::tarmak::etcd_ssl_dir}/${::tarmak::etcd_k8s_events_ca_name}-key.pem",
    tls_ca_path              => "${::tarmak::etcd_ssl_dir}/${::tarmak::etcd_k8s_events_ca_name}-ca.pem",
    systemd_after            => delete_undef_values([$::tarmak::etcd_mount_unit]),
    systemd_requires         => delete_undef_values([$::tarmak::etcd_mount_unit]),
  }
  etcd::instance{'overlay':
    version                  => $::tarmak::etcd_overlay_version,
    nodename                 => $nodename,
    members                  => $::tarmak::etcd_instances,
    initial_cluster          => $::tarmak::_etcd_cluster,
    advertise_client_network => $::tarmak::etcd_advertise_client_network,
    client_port              => $::tarmak::etcd_overlay_client_port,
    peer_port                => $::tarmak::etcd_overlay_peer_port,
    tls                      => true,
    tls_cert_path            => "${::tarmak::etcd_ssl_dir}/${::tarmak::etcd_overlay_ca_name}.pem",
    tls_key_path             => "${::tarmak::etcd_ssl_dir}/${::tarmak::etcd_overlay_ca_name}-key.pem",
    tls_ca_path              => "${::tarmak::etcd_ssl_dir}/${::tarmak::etcd_overlay_ca_name}-ca.pem",
    systemd_after            => delete_undef_values([$::tarmak::etcd_mount_unit]),
    systemd_requires         => delete_undef_values([$::tarmak::etcd_mount_unit]),
  }
}
