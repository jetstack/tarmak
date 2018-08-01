include ::tarmak
define consul::consul_backup_service (
    String $region,
    String $backup_bucket_prefix,
    String $consul_master_token,
)
{
    require consul

    $service_name = 'consul-backup'

    file { "${::consul::systemd_dir}/${service_name}.service":
        ensure  => file,
        content => template('consul/consul-backup.service.erb'),
        notify  => Exec["${service_name}-systemctl-daemon-reload"],
        mode    => '0644'
    }
    ~> exec { "${service_name}-systemctl-daemon-reload":
        command     => 'systemctl daemon-reload',
        refreshonly => true,
        path        => $::consul::path,
    }
    -> service { "${service_name}.service":
        ensure => 'running',
        enable => true,
    }

    file { "${vault_client::systemd_dir}/${service_name}.timer":
        ensure  => file,
        content => template('vault_client/cert.timer.erb'),
        notify  => Exec["${service_name}-systemctl-daemon-reload"],
    }
    ~> service { "${service_name}.timer":
        ensure  => 'running',
        enable  => true,
        require => Exec["${service_name}-systemctl-daemon-reload"],
    }
}
