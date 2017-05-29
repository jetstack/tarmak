# class kubernetes::kubelet
class kubernetes::proxy(
  $ca_file = undef,
  $cert_file = undef,
  $key_file = undef,
){
  require ::kubernetes

  $service_name = 'kube-proxy'

  $kubeconfig_path = "${::kubernetes::config_dir}/kubeconfig-proxy"
  file{$kubeconfig_path:
    ensure  => file,
    mode    => '0640',
    owner   => 'root',
    group   => $kubernetes::group,
    content => template('kubernetes/kubeconfig.erb'),
    notify  => Service["${service_name}.service"],
  }

  kubernetes::symlink{'proxy':} ->
  file{"${::kubernetes::systemd_dir}/${service_name}.service":
    ensure  => file,
    mode    => '0644',
    owner   => 'root',
    group   => 'root',
    content => template("kubernetes/${service_name}.service.erb"),
    notify  => Service["${service_name}.service"],
  } ~>
  exec { "${service_name}-daemon-reload":
    command     => 'systemctl daemon-reload',
    path        => $::kubernetes::path,
    refreshonly => true,
  } ->
  service{ "${service_name}.service":
    ensure => running,
    enable => true,
  }
}
