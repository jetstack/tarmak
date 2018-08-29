class consul::service(
  $consul_encrypt = $::consul::consul_encrypt,
  $fqdn = $consul::fqdn,
  $private_ip = $consul::private_ip,
  $consul_master_token = $consul::consul_master_token,
  $region = $consul::region,
  $instance_count = $consul::instance_count,
  $environment = $consul::environment,
  $backup_bucket_prefix = $consul::backup_bucket_prefix,
  $backup_schedule = $consul::backup_schedule,
  $volume_id = $consul::volume_id,
)
{

  $service_name = 'consul'
  $backup_service_name = 'consul-backup'

  $token_file_path = "${consul::config_dir}/master-token"

  $bin_path = $consul::bin_path
  $config_path = $consul::config_path

  $user = $consul::user
  $group = $consul::group
  $data_dir = $consul::data_dir

  exec { "${service_name}-systemctl-daemon-reload":
    command     => '/bin/systemctl daemon-reload',
    refreshonly => true,
    path        => defined('$::path') ? {
      default => '/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/opt/bin',
      true    => $::path
    },
  }

  # install consul exporter if enabled
  if $consul::exporter_enabled {
    $exporter_bin_path = $consul::exporter_bin_path
    file { "${::consul::systemd_dir}/${service_name}-exporter.service":
      ensure  => file,
      content => template('consul/consul-exporter.service.erb'),
      notify  => Exec["${service_name}-systemctl-daemon-reload"]
    }
    ~> service { "${service_name}-exporter.service":
      ensure  => running,
      enable  => true,
      require => [
        Exec["${service_name}-systemctl-daemon-reload"],
        Service["${service_name}.service"],
        ],
    }

    if defined('$consul::consul_master_token') {
      File[$token_file_path] ~> Service["${service_name}-exporter.service"]
    }
  }

  file { "${::consul::systemd_dir}/${service_name}.service":
    ensure  => file,
    content => template('consul/consul.service.erb'),
    notify  => Exec["${service_name}-systemctl-daemon-reload"]
  }
  ~> service { "${service_name}.service":
    ensure  => running,
    enable  => true,
    require => [
      Exec["${service_name}-systemctl-daemon-reload"],
      ],
  }

  $disks = aws_ebs::disks()
  case $disks.length {
    0: {$ebs_device = ''}
    1: {$ebs_device = $disks[0]}
    default: {$ebs_device = $disks[1]}
  }

  if $::consul::cloud_provider == 'aws' {
    class{'::aws_ebs':
      bin_dir     => $::consul::_dest_dir,
      systemd_dir => $::consul::systemd_dir,
    }
    aws_ebs::mount{'consul':
      volume_id => $volume_id,
      device    => $ebs_device,
      dest_path => $data_dir,
    }
  }

  file { "${::consul::systemd_dir}/${backup_service_name}.service":
    ensure  => file,
    content => template('consul/consul-backup.service.erb'),
    notify  => Exec["${service_name}-systemctl-daemon-reload"],
    mode    => '0644'
  }

  file { "${consul::systemd_dir}/${backup_service_name}.timer":
    ensure  => file,
    content => template('consul/consul-backup.timer.erb'),
    notify  => Exec["${service_name}-systemctl-daemon-reload"],
  }
  ~> service { "${backup_service_name}.timer":
    ensure  => 'running',
    enable  => true,
    require => Exec["${service_name}-systemctl-daemon-reload"],
  }
}
