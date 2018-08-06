include ::tarmak
define consul::attach_ebs_volume_service (
    String $region,
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
        command     => '/bin/systemctl daemon-reload',
        refreshonly => true,
        path        => defined('$::path') ? {
          default => '/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/opt/bin',
          true    => $::path
        },
    }
    -> service { "${service_name}.service":
        ensure => 'running',
        enable => true,
    }
}