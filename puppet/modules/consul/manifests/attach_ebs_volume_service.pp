include ::tarmak
define consul::attach_ebs_volume_service (
    String $region,
    String $volume,
    String $volume_id,
)
{
    require consul

    $service_name = 'attach-ebs-volume'

    file { "${::consul::systemd_dir}/${service_name}.service":
        ensure  => file,
        content => template('consul/attach-ebs-volume.service.erb'),
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
