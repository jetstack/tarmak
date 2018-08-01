include ::tarmak
define consul::ensure_ebs_volume_service (
)
{
    require consul

    $service_name = 'ensure-ebs-volume-formatted'

    file { "${::consul::systemd_dir}/${service_name}.service":
        ensure  => file,
        content => template('consul/ensure-ebs-volume-formatted.service.erb'),
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
}
