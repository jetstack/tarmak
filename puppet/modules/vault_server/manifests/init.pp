class vault_server (
    $region,
    $vault_tls_cert_path,
    $vault_tls_ca_path,
    $vault_tls_key_path,
    $vault_unsealer_kms_key_id,
    $vault_unsealer_ssm_key_prefix,
    String $init_role = '',
    String $version = $::vault_server::params::version,
    String $bin_dir = $::vault_server::params::bin_dir,
    String $local_bin_dir = $::vault_server::params::local_bin_dir,
    String $download_dir = $::vault_server::params::download_dir,
    String $dest_dir = $::vault_server::params::dest_dir,
    String $server_url = $::vault_server::params::server_url,
    String $systemd_dir = $::vault_server::params::systemd_dir,
) inherits ::vault_server::params {

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

    $_dest_dir = "${dest_dir}/${::vault_server::params::app_name}-${version}"
    $bin_path = "${_dest_dir}/${::vault_server::params::app_name}"
    $link_path = "/usr/local/bin/${::vault_server::params::app_name}"

    file { '/etc/vault':
        ensure => 'directory',
        mode   => '0777',
    }

    file { '/var/lib/vault':
        ensure => 'directory',
        mode   => '0777',
    }

    user { 'vault':
        ensure => 'present',
        system => true,
        home   => '/var/lib/vault',
    }

    Class['::airworthy']
    -> class { '::airworthy::install': }
    -> class { '::vault_server::install': }
    ~> class { '::vault_server::service': }
}
