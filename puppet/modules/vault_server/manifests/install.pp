class vault_server::install(
    String $user = 'root',
    String $group = 'root',
)
{
    include ::archive

    # install unzip if necessary
    ensure_resource('package', 'unzip', {})

    $zip_path = "${vault_server::_dest_dir}/vault.zip"
    $unsealer_script_name = 'download-vault-unsealer'
    $vault_script_name = 'vault'

    Package['unzip']
    -> Archive[$zip_path]

    file {$vault_server::_dest_dir:
        ensure => directory,
        mode   => '0755',
    }
    -> archive {$zip_path:
        ensure           => present,
        extract          => true,
        source           => $vault_server::download_url,
        signature_binary => $vault_server::signature_url,
        sha256sums       => $vault_server::sha256sums_url,
        provider         => 'airworthy',
        extract_path     => $vault_server::_dest_dir,
    }
    -> file {$vault_server::bin_path:
        mode   => '0755',
    }
    -> file {$vault_server::link_path:
        ensure => link,
        target => $vault_server::bin_path,
    }

    file { "${vault_server::local_bin_dir}/${unsealer_script_name}.sh":
        ensure  => file,
        content => file('vault_server/download-vault-unsealer.sh'),
        owner   => $user,
        group   => $group,
        mode    => '0755',
    }
    ~> exec { "${unsealer_script_name}-script-run":
        command => "${::vault_server::local_bin_dir}/${unsealer_script_name}.sh",
        path    => $vault_server::path,
    }

    file { "/etc/profile.d/${vault_script_name}.sh":
        ensure  => file,
        content => file('vault_server/vault.sh'),
        owner   => $user,
        group   => $group,
        mode    => '0644',
    }
}
