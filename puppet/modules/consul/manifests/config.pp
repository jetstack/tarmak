class consul::config (
    $consul_encrypt = $consul::consul_encrypt,
    $private_ip = $consul::private_ip,
    $consul_master_token = $consul::consul_master_token,
    $region = $consul::region,
    $instance_count = $consul::instance_count,
    $environment = $consul::environment,
)
{

    if $consul::ca_file != undef and $consul::cert_file != undef and $consul::key_file != undef {
        $enable_tls = true
    }
    else {
        $enable_tls = false
    }

    if $consul::advertise_network == undef {
        if defined('$::ipaddress') {
            $advertise_addr = $::ipaddress
        } else {
            $advertise_addr = '127.0.0.1'
        }
    } else {
        $advertise_addr = get_ipaddress_in_network($consul::advertise_network)
    }


    # build config hash
    $config = {
        acl_default_policy  => defined('$consul_master_token') ? {
            true    => "\"${consul::acl_default_policy}\"",
            default =>  undef,
        },
        acl_down_policy     => defined('$consul_master_token') ? {
            true    => "\"${consul::acl_down_policy}\"",
            default =>  undef,
        },
        acl_master_token    => defined('$consul_master_token') ? {
            true    => "\"${consul_master_token}\"",
            default =>  undef,
        },
        acl_agent_token    => defined('$consul_master_token') ? {
            true    => "\"${consul_master_token}\"",
            default =>  undef,
        },
        acl_datacenter      => defined('$consul_master_token') ? {
            true    => "\"${consul::datacenter}\"",
            default =>  undef,
        },
        datacenter          => "\"${consul::datacenter}\"",
        log_level           => "\"${consul::log_level}\"",
        client_addr         => "\"${consul::client_addr}\"",
        bind_addr           => "\"${consul::bind_addr}\"",
        advertise_addr      => "\"${advertise_addr}\"",
        encrypt             => defined('$consul_encrypt') ? {
            true    => "\"${consul_encrypt}\"",
            default =>  undef,
        },
        bootstrap_expect    => defined('$consul::consul_bootstrap_expect') ? {
            true    =>  $consul::consul_bootstrap_expect.scanf('%i')[0],
            default =>  1,
        },
        server              => $consul::server,
        disable_remote_exec => true,
        retry_join          => $consul::cloud_provider ? {
            'aws'   => "[\"provider=aws tag_key=VaultCluster tag_value=${environment}\"]",
            default => $consul::retry_join,
        },
        ca_file             => "\"${consul::ca_file}\"",
        cert_file           => "\"${consul::cert_file}\"",
        key_file            => "\"${consul::key_file}\"",
        verify_outgoing     => $enable_tls,
        verify_incoming     => $enable_tls,
    }

    file { $consul::config_dir:
        ensure => directory,
        owner  => $consul::user,
        group  => $consul::group,
        mode   => '0750',
    }
    -> file { $consul::config_path:
        ensure  => file,
        owner   => $consul::user,
        group   => $consul::group,
        mode    => '0600',
        content => epp('consul/consul.json.epp', {
            config => $config,
        })
    }

    # write master token to vault
    if defined('$consul::consul_master_token') {
        $token_file_path = "${consul::config_dir}/master-token"
        file {$token_file_path:
            ensure  => file,
            content => "CONSUL_HTTP_TOKEN=${consul::consul_master_token}",
            owner   => $consul::user,
            group   => $consul::group,
            mode    => '0600',
        }
    }
}
