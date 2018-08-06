include ::tarmak
define consul::consul_backup_service (
    String $region,
    String $backup_bucket_prefix,
    String $backup_schedule,
    String $consul_master_token,
)
{
    require consul

    $service_name = 'consul-backup'

    file { "/usr/local/bin/${service_name}.sh":
        ensure  => file,
        content => file('consul/consul-backup.sh'),
        mode    => '0755'
    }

    file { "${::consul::systemd_dir}/${service_name}.service":
        ensure  => file,
        content => template('consul/consul-backup.service.erb'),
        notify  => Exec["${service_name}-systemctl-daemon-reload"],
        mode    => '0644'
    }
    ~> exec { "${service_name}-systemctl-daemon-reload":
        command     => '/bin/systemctl daemon-reload',
        refreshonly => true,
        path        => $::consul::path,
    }

    file { "${consul::systemd_dir}/${service_name}.timer":
        ensure  => file,
        content => template('consul/consul-backup.timer.erb'),
        notify  => Exec["${service_name}-systemctl-daemon-reload"],
    }
    ~> service { "${service_name}.timer":
        ensure  => 'running',
        enable  => true,
        require => Exec["${service_name}-systemctl-daemon-reload"],
    }
}
