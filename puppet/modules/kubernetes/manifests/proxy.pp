# class kubernetes::kubelet
class kubernetes::proxy(
  Optional[String] $ca_file = undef,
  Optional[String] $cert_file = undef,
  Optional[String] $key_file = undef,
  Array[String] $systemd_wants = [],
  Array[String] $systemd_requires = [],
  Array[String] $systemd_after = [],
  Array[String] $systemd_before = [],
){
  require ::kubernetes

  $service_name = 'kube-proxy'

  $_systemd_wants = $systemd_wants
  $_systemd_after = $systemd_after
  $_systemd_requires = $systemd_after
  $_systemd_before = $systemd_before

  # ensure ipvsadm and contrack installed (for kube-proxy)
  $conntrack_package_name = 'conntrack'
  ensure_resource('package', [$conntrack_package_name,'ipvsadm'],{
    ensure => 'present',
  })

  $kubeconfig_path = "${::kubernetes::config_dir}/kubeconfig-proxy"
  file{$kubeconfig_path:
    ensure  => file,
    mode    => '0640',
    owner   => 'root',
    group   => $kubernetes::group,
    content => template('kubernetes/kubeconfig.erb'),
    notify  => Service["${service_name}.service"],
  }

  kubernetes::symlink{'proxy':}
  -> file{"${::kubernetes::systemd_dir}/${service_name}.service":
    ensure  => file,
    mode    => '0644',
    owner   => 'root',
    group   => 'root',
    content => template("kubernetes/${service_name}.service.erb"),
    notify  => Service["${service_name}.service"],
  }
  ~> exec { "${service_name}-daemon-reload":
    command     => 'systemctl daemon-reload',
    path        => $::kubernetes::path,
    refreshonly => true,
  }
  -> service{ "${service_name}.service":
    ensure  => running,
    enable  => true,
    require => Package[$conntrack_package_name],
  }
}
