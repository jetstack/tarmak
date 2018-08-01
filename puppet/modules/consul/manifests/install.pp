class consul::install(
    String $download_url = 'https://releases.hashicorp.com/consul/#VERSION#/consul_#VERSION#_linux_amd64.zip',
    String $sha256sums_url = 'https://releases.hashicorp.com/consul/#VERSION#/consul_#VERSION#_SHA256SUMS',
    String $signature_url = 'https://releases.hashicorp.com/consul/#VERSION#/consul_#VERSION#_SHA256SUMS.sig',
    String $exporter_download_url = 'https://github.com/prometheus/consul_exporter/releases/download/v#VERSION#/consul_exporter-#VERSION#.linux-amd64.tar.gz',
    String $exporter_signature_url = 'https://releases.tarmak.io/signatures/consul_exporter/#VERSION#/consul_exporter-#VERSION#.linux-amd64.tar.gz.asc',
)
{
    include ::archive


    $nologin = $::osfamily ? {
        'RedHat' => '/sbin/nologin',
        'Debian' => '/usr/sbin/nologin',
        default  => '/usr/sbin/nologin',
    }

    group { $::consul::group:
        ensure => present,
        gid    => $::consul::gid,
    }
    -> user { $::consul::user:
        ensure => present,
        uid    => $::consul::uid,
        shell  => $nologin,
        home   => $::consul::data_dir,
    }

    $version = $name

    $_download_url = regsubst(
        $download_url,
        '#VERSION#',
        $::consul::version,
        'G'
    )

    $_sha256sums_url = regsubst(
        $sha256sums_url,
        '#VERSION#',
        $::consul::version,
        'G'
    )

    $_signature_url = regsubst(
        $signature_url,
        '#VERSION#',
        $::consul::version,
        'G'
    )

    # install unzip if necessary
    ensure_resource('package', 'unzip', {})

    $zip_path = "${::consul::_dest_dir}/consul.zip"

    Package['unzip']
    -> Archive[$zip_path]

    file {$consul::_dest_dir:
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
        extract_path     => $::consul::_dest_dir,
    }
    -> file {$::consul::bin_path:
        mode   => '0755',
    }
    -> file {$::consul::link_path:
        ensure => link,
        target => $::consul::bin_path,
    }

    # install consul exporter if enabled
    if $::consul::exporter_enabled {
        $exporter_tar_path = "${::consul::exporter_dest_dir}/consul-exporter.tar.gz"

        $_exporter_download_url = regsubst(
            $exporter_download_url,
            '#VERSION#',
            $::consul::exporter_version,
            'G'
        )

        $_exporter_signature_url = regsubst(
            $exporter_signature_url,
            '#VERSION#',
            $::consul::exporter_version,
            'G'
        )

        file {$::consul::exporter_dest_dir:
            ensure => directory,
            mode   => '0755',
        }
        -> archive {$exporter_tar_path:
            ensure            => present,
            extract           => true,
            source            => $_exporter_download_url,
            signature_armored => $_exporter_signature_url,
            provider          => 'airworthy',
            extract_path      => $::consul::exporter_dest_dir,
            extract_command   => 'tar xfz %s --strip-components=1'
        }
    }
    else {
        file {$::consul::exporter_dest_dir:
            ensure => absent,
        }
    }


    consul::consul{'consul':
        fqdn                => $::consul::fqdn,
        private_ip          => $::consul::private_ip,
        consul_master_token => $::consul::consul_master_token,
        region              => $::consul::region,
        instance_count      => $::consul::instance_count,
        environment         => $::consul::environment,
        consul_encrypt      => $::consul::consul_encrypt,
    }
}
