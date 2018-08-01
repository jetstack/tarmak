class vault_server (
    String $version = $::vault_server::params::version,
    String $bin_dir = $::vault_server::params::bin_dir,
    String $local_bin_dir = $::vault_server::params::local_bin_dir,
    String $download_dir = $::vault_server::params::download_dir,
    String $dest_dir = $::vault_server::params::dest_dir,
    String $server_url = $::vault_server::params::server_url,
    String $systemd_dir = $::vault_server::params::systemd_dir,
    String $init_token = "",
    String $init_role = "",
    String $token = "",
    String $ca_cert_path = "",
    $region,
    $vault_tls_cert_path,
    $vault_tls_ca_path,
    $vault_tls_key_path,
    $vault_unsealer_kms_key_id,
    $vault_unsealer_ssm_key_prefix,
) inherits ::vault_server::params {

    # verify inputs

    ## only one of init_token or token needs to exist
    #if $init_token == undef and $token == undef {
    #  fail('You must provide at least one of $init_token or $token.')
    #}
    #if $init_token != undef and $token != undef {
    #  fail('You must provide either $init_token or $token.')
    #}

    # paths
    $path = defined('$::path') ? {
        default => '/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/opt/bin',
        true    => $::path,
    }

    ## build download URL
    $download_url = regsubst(
        $::vault_server::params::download_url,
        '#VERSION#',
        $version,
        'G'
    )

    # token path
    $config_path = "${::vault_server::config_dir}/config"

    # token path
    $token_path = "${::vault_server::config_dir}/token"

    # init_token path
    $init_token_path = "${::vault_server::config_dir}/init-token"

    $_dest_dir = "${dest_dir}/${::vault_server::params::app_name}-${version}"

    user { 'vault':
        ensure => 'present',
        system => true,
        home   => '/var/lib/vault',
    }

    vault_server::assets_service{'assets_service':
        vault_tls_cert_path => $vault_tls_cert_path,
        vault_tls_ca_path   => $vault_tls_ca_path,
        vault_tls_key_path  => $vault_tls_key_path,
    }

    vault_server::download_unsealer{'download_unsealer':
    }

    vault_server::vault{'vault_script':
    }

    vault_server::unsealer_service{'unsealer_service':
        region                        => $region,
        vault_unsealer_kms_key_id     => $vault_unsealer_kms_key_id,
        vault_unsealer_ssm_key_prefix => $vault_unsealer_ssm_key_prefix,
    }

    vault_server::vault_service{'vault_service':
        region => $region,
    }

    class { '::vault_server::config': }
    -> Class['::vault_server']
}
