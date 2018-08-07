class vault_server::install(
    String $download_url = 'https://releases.hashicorp.com/vault/#VERSION#/vault_#VERSION#_linux_amd64.zip',
    String $sha256sums_url = 'https://releases.hashicorp.com/vault/#VERSION#/vault_#VERSION#_SHA256SUMS',
    String $signature_url = 'https://releases.hashicorp.com/vault/#VERSION#/vault_#VERSION#_SHA256SUMS.sig',
    String $user = 'root',
    String $group = 'root',
)
{
    include ::archive

    $_download_url = regsubst(
        $download_url,
        '#VERSION#',
        $::vault_server::version,
        'G'
    )

    $_sha256sums_url = regsubst(
        $sha256sums_url,
        '#VERSION#',
        $::vault_server::version,
        'G'
    )

    $_signature_url = regsubst(
        $signature_url,
        '#VERSION#',
        $::vault_server::version,
        'G'
    )


    # install unzip if necessary
    ensure_resource('package', 'unzip', {})

    $zip_path = "${::vault_server::_dest_dir}/vault.zip"

    Package['unzip']
    -> Archive[$zip_path]

    file {$vault_server::_dest_dir:
        ensure => directory,
        mode   => '0755',
    }
    -> archive {$zip_path:
        ensure           => present,
        extract          => true,
        source           => $_download_url,
        signature_binary => $_signature_url,
        sha256sums       => $_sha256sums_url,
        provider         => 'airworthy',
        extract_path     => $::vault_server::_dest_dir,
    }
    -> file {$::vault_server::bin_path:
        mode   => '0755',
    }
    -> file {$::vault_server::link_path:
        ensure => link,
        target => $::vault_server::bin_path,
    }


    $unsealer_script_name = 'download-vault-unsealer'

    file { "${::vault_server::local_bin_dir}/${unsealer_script_name}.sh":
        ensure  => file,
        content => file('vault_server/download-vault-unsealer.sh'),
        owner   => $user,
        group   => $group,
        mode    => '0755',
    }
    ~> exec { "${unsealer_script_name}-script-run":
        command => "${::vault_server::local_bin_dir}/${unsealer_script_name}.sh",
        path    => $::vault_server::path,
    }


    $vault_script_name = 'vault'

    file { "/etc/profile.d/${vault_script_name}.sh":
        ensure  => file,
        content => file('vault_server/vault.sh'),
        owner   => $user,
        group   => $group,
        mode    => '0644',
    }
}
