include ::tarmak
define consul::consul(
    String $fqdn,
    String $private_ip,
    String $consul_master_token,
    String $region,
    String $sonul_encrypt,
    Integer $instance_count,
    String $environment,
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

    file { "${::consul::consul_config_dir}/${json_name}.json":
        ensure  => file,
        content => template('consul/consul.json.erb'),
        mode    => '0600'
    }

    file { "${::consul::vault_config_dir}/${hcl_name}.hcl":
        ensure  => file,
        content => template('consul/vault.hcl.erb'),
        mode    => '0600'
    }

    file { "${::consul::bin_dir}/download-vault-consul.sh":
        ensure  => file,
        content => template('consul/download-vault-consul.sh.erb'),
        mode    => '0755'
    }

    file { "${::vault_client::systemd_dir}/${service_name}.service":
        ensure  => file,
        content => template('vault_client/consul.service.erb'),
        notify  => Service["${service_name}.service"],
        mode    => '0644'
    }
    ~> exec { "${service_name}-systemctl-daemon-reload":
        command     => 'systemctl daemon-reload',
        refreshonly => true,
        path        => $::vault_client::path,
    }
    -> service { "${service_name}.service":
        ensure => 'running',
        enable => true,
    }
}
