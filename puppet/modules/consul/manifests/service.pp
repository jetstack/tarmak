class consul::service(
  $systemd_wants = [],
  $systemd_requires = [],
  $systemd_after = [],
  $systemd_before = [],
  $private_ip = $consul::private_ip,
)
{
  $service_name = 'consul'
  $mount_name = 'var-lib-consul'

  $_systemd_wants = $systemd_wants
  $_systemd_requires = $systemd_requires
  $_systemd_after = ['network.target'] + $systemd_after
  $_systemd_before = $systemd_before

  $bin_path = $::consul::bin_path
  $config_path = $::consul::config_path

  $user = $::consul::user
  $group = $::consul::group
  $data_dir = $::consul::data_dir

  exec { "${service_name}-systemctl-daemon-reload":
    command     => 'systemctl daemon-reload',
    refreshonly => true,
    path        => defined('$::path') ? {
      default => '/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/opt/bin',
      true    => $::path
    },
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

  # write master token to vault
  #if defined('$::consul::consul_master_token') {
    $token_file_path = "${::consul::config_dir}/master-token"
    file {$token_file_path:
      ensure  => file,
      content => "CONSUL_HTTP_TOKEN=${::consul::consul_master_token}",
      owner   => $::consul::user,
      group   => $::consul::group,
      mode    => '0600',
    }
    #}

  # install consul exporter if enabled
  if $::consul::exporter_enabled {
    $exporter_bin_path = $::consul::exporter_bin_path
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

    if defined('$::consul::consul_master_token') {
      File[$token_file_path] ~> Service["${service_name}-exporter.service"]
    }
  }
}
