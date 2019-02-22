define aws_es_proxy::instance (
  String $dest_address,
  Boolean $tls = true,
  Integer $dest_port = 9200,
  Integer $listen_port = 9200,
  Enum['present', 'absent'] $ensure = 'present'
){
  include ::aws_es_proxy

  $service_name = "aws-es-proxy-${title}"

  $proxy_path = $::aws_es_proxy::proxy_path
  $path = $::aws_es_proxy::path

  if $tls {
    if $dest_port == 443 {
      $endpoint = "https://${dest_address}"
    } else {
      $endpoint = "https://${dest_address}:${dest_port}"
    }
  } else {
    if $dest_port == 80 {
      $endpoint = "http://${dest_address}"
    } else {
      $endpoint = "http://${dest_address}:${dest_port}"
    }
  }

  if $ensure == 'present' {
    $service_ensure = 'running'
    $service_enable = true
  } else {
    $service_ensure = 'stopped'
    $service_enable = false
  }

  exec { "${service_name}-systemctl-daemon-reload":
    command     => 'systemctl daemon-reload',
    refreshonly => true,
    path        => $path,
  }

  file{ "/etc/systemd/system/${service_name}.service":
    ensure  => file,
    mode    => '0644',
    owner   => 'root',
    group   => 'root',
    content => template('aws_es_proxy/aws-es-proxy.service.erb'),
    notify  => Exec["${service_name}-systemctl-daemon-reload"]
  }
  ~> service { "${service_name}.service":
    ensure     => $service_ensure,
    enable     => $service_enable,
    hasstatus  => true,
    hasrestart => true,
    subscribe  => Class['aws_es_proxy::install']
  }
}
