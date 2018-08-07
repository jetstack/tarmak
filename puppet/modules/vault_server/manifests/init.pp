class vault_server (
    $region,
    $vault_tls_cert_path,
    $vault_tls_ca_path,
    $vault_tls_key_path,
    $vault_unsealer_kms_key_id,
    $vault_unsealer_ssm_key_prefix,

    String $app_name = $vault_server::params::app_name,
    String $version = $vault_server::params::version,
    String $bin_dir = $vault_server::params::bin_dir,
    String $local_bin_dir = $vault_server::params::local_bin_dir,
    String $dest_dir = $vault_server::params::dest_dir,
    String $config_dir = $vault_server::params::config_dir,
    String $lib_dir = $vault_server::params::lib_dir,
    String $download_dir = $vault_server::params::download_dir,
    String $_download_url = $vault_server::params::download_url,
    String $_sha256sums_url = $vault_server::params::sha256sums_url,
    String $_signature_url = $vault_server::params::signature_url,
    String $systemd_dir = $vault_server::params::systemd_dir,
) inherits ::vault_server::params {

    include ::archive
    include ::airworthy

    # paths
    $path = defined('$::path') ? {
        default => '/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/opt/bin',
        true    => $::path,
    }

    ## build download URL
    $download_url = regsubst(
        $_download_url,
        '#VERSION#',
        $version,
        'G'
    )

    $sha256sums_url = regsubst(
        $_sha256sums_url,
        '#VERSION#',
        $version,
        'G'
    )

    $signature_url = regsubst(
        $_signature_url,
        '#VERSION#',
        $version,
        'G'
    )

    $_dest_dir = "${dest_dir}/${app_name}-${version}"
    $bin_path = "${_dest_dir}/${app_name}"
    $link_path = "${local_bin_dir}/${app_name}"

    file { $config_dir:
        ensure => 'directory',
        mode   => '0777',
    }

    file { $lib_dir:
        ensure => 'directory',
        mode   => '0777',
    }

    user { $app_name:
        ensure => 'present',
        system => true,
        home   => $lib_dir,
    }

    Class['::airworthy']
    -> class { '::airworthy::install': }
    -> class { '::vault_server::install': }
    ~> class { '::vault_server::service': }
}
