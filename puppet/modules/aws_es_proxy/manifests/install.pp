class aws_es_proxy::install {

  ensure_resource('package', ['curl'],{
    ensure => present
  })

  Package['curl']
  -> file { $::aws_es_proxy::_dest_dir:
    ensure => directory,
    mode   => '0755',
  }
  -> exec {"aws-es-proxy-${::aws_es_proxy::version}-download":
    command => "curl -sL -o ${::aws_es_proxy::proxy_path} ${::aws_es_proxy::_download_url}",
    creates => $::aws_es_proxy::proxy_path,
    path    => ['/usr/bin/', '/bin'],
  }
  -> file {$::aws_es_proxy::proxy_path:
    ensure => file,
    mode   => '0755',
    owner  => 'root',
    group  => 'root',
  }

}

