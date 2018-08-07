class vault_server::service (
    $region = $vault_server::region,
    $vault_tls_cert_path = $vault_server::vault_tls_cert_path,
    $vault_tls_ca_path = $vault_server::vault_tls_ca_path,
    $vault_tls_key_path = $vault_server::vault_tls_key_path,
    $vault_unsealer_kms_key_id = $vault_server::vault_unsealer_kms_key_id,
    $vault_unsealer_ssm_key_prefix = $vault_server::vault_unsealer_ssm_key_prefix,
    $user = 'root',
    $group = 'root',
    $assets_service_name = 'vault-assets',
    $unsealer_service_name = 'vault-unsealer',
    $vault_service_name = 'vault',
)
{

    file { "${::vault_server::systemd_dir}/${assets_service_name}.service":
        ensure  => file,
        content => template('vault_server/vault-assets.service.erb'),
        notify  => Service["${assets_service_name}.service"],
        owner   => $user,
        group   => $group,
        mode    => '0644',
    }
    ~> exec { "${assets_service_name}-systemctl-daemon-reload":
        command     => 'systemctl daemon-reload',
        refreshonly => true,
        path        => $vault_server::path,
    }
    -> service { "${assets_service_name}.service":
        ensure => 'stopped',
        enable => false,
    }


    file { "${::vault_server::systemd_dir}/${unsealer_service_name}.service":
        ensure  => file,
        content => template('vault_server/vault-unsealer.service.erb'),
        notify  => Service["${unsealer_service_name}.service"],
        owner   => $user,
        group   => $group,
        mode    => '0644',
    }
    ~> exec { "${unsealer_service_name}-systemctl-daemon-reload":
        command     => 'systemctl daemon-reload',
        refreshonly => true,
        path        => $vault_server::path,
    }
    -> service { "${unsealer_service_name}.service":
        ensure => 'running',
        enable => true,
    }

    file { "${::vault_server::systemd_dir}/${vault_service_name}.service":
        ensure  => file,
        content => template('vault_server/vault.service.erb'),
        notify  => Service["${vault_service_name}.service"],
        owner   => $user,
        group   => $group,
        mode    => '0644',
    }
    ~> exec { "${vault_service_name}-systemctl-daemon-reload":
        command     => 'systemctl daemon-reload',
        refreshonly => true,
        path        => $vault_server::path,
    }
    -> service { "${vault_service_name}.service":
        ensure => 'running',
        enable => true,
    }
}
