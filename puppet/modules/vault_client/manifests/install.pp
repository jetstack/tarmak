class vault_client::install {
    $vault_helper_path = "${::vault_client::dest_dir}/vault-helper-${vault_client::version}"
    $vault_helper_bin = '/opt/bin/vault-helper'

    file { $::vault_client::dest_dir:
        ensure => directory,
        mode   => '0755',
    }
    -> file { $vault_helper_path:
        ensure => directory,
        mode   => '0755',
    }
    -> exec {"vault-helper-${vault_client::version}-download":
        command => "curl -sL ${::vault_client::download_url} -o ${vault_helper_path}/vault-helper",
        creates => $vault_helper_path,
        path    => ['/usr/bin', '/bin'],
    }
    -> file { "${vault_helper_path}/vault-helper":
        ensure => 'file',
        mode   => '0755',
        owner  => 'root',
        group  => 'root',
    }
    -> file { $vault_helper_bin:
        ensure => 'link',
        mode   => '0755',
        owner  => 'root',
        group  => 'root',
        target => "${vault_helper_path}/vault-helper",
    }
}
