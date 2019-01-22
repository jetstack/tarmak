class consul::service(
  String $fqdn = $consul::fqdn,
  String $private_ip = $consul::private_ip,
  String $region = $consul::region,
  String $environment = $consul::environment,
  String $backup_bucket_prefix = $consul::backup_bucket_prefix,
  String $backup_schedule = $consul::backup_schedule,
  Array[String] $_systemd_wants = $consul::systemd_wants,
  Array[String] $_systemd_before = $consul::systemd_before,
  Integer $consul_bootstrap_expect = $consul::_consul_bootstrap_expect
)
{

  $_systemd_after = ['network.target', 'var-lib-consul.mount'] + $consul::systemd_after
  $_systemd_requires = ['var-lib-consul.mount'] + $consul::systemd_requires

  $service_name = 'consul'
  $backup_service_name = 'consul-backup'

  $token_file_path = "${consul::config_dir}/master-token"

  $bin_path = $consul::bin_path
  $config_path = $consul::config_path

  $consul_encrypt = $consul::_consul_encrypt
  $consul_master_token = $consul::_consul_master_token

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

    if defined('$consul_master_token') {
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
    require => Exec["${service_name}-systemctl-daemon-reload"],
  }


  file { "${::consul::systemd_dir}/${backup_service_name}.service":
    ensure  => file,
    content => template('consul/consul-backup.service.erb'),
    notify  => Exec["${service_name}-systemctl-daemon-reload"],
    mode    => '0644'
  }
  ~> service { "${backup_service_name}.service":
    ensure  => 'stopped',
    enable  => false,
    require => Exec["${service_name}-systemctl-daemon-reload"],
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
