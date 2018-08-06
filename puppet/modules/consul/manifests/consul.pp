define consul::consul(
    String $fqdn,
    String $private_ip,
    String $consul_master_token,
    String $region,
    String $instance_count,
    String $environment,
    String $consul_encrypt
)
{
    require consul

    $mount_name = 'var-lib-consul'
    $service_name = 'consul'
    $json_name = 'consul'
    $hcl_name = 'vault'

    file { "${::consul::systemd_dir}/${mount_name}.mount":
        ensure  => file,
        content => template('consul/var-lib-consul.mount.erb'),
        notify  => Exec["${service_name}-systemctl-daemon-reload"],
        mode    => '0644'
    }
    ~> exec { "${mount_name}-mount":
        command => "/bin/systemctl enable ${mount_name}.mount",
        path    => $::consul::path,
    }

    file { "${::consul::vault_config_dir}/${hcl_name}.hcl":
        ensure  => file,
        content => template('consul/vault.hcl.erb'),
        mode    => '0600'
    }
}
