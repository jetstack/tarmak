# class kubernetes::proxy
class kubernetes::proxy(
  String $service_ensure = 'running',
  Optional[String] $ca_file = undef,
  Optional[String] $cert_file = undef,
  Optional[String] $key_file = undef,
  Array[String] $systemd_wants = [],
  Array[String] $systemd_requires = [],
  Array[String] $systemd_after = [],
  Array[String] $systemd_before = [],
  Hash[String,Boolean] $feature_gates = {},
  String $config_file = "${::kubernetes::params::config_dir}/kube-proxy-config.yaml",
) inherits kubernetes::params{
  require ::kubernetes

  $service_name = 'kube-proxy'

  $_systemd_wants = $systemd_wants
  $_systemd_after = $systemd_after
  $_systemd_requires = $systemd_after
  $_systemd_before = $systemd_before

  if $feature_gates == {} {
    $_feature_gates = $::kubernetes::_enable_pod_priority ? {
      true    =>  {'PodPriority' => true},
      default => undef,
    }
  } else {
    $_feature_gates = $feature_gates
  }

  $post_1_11 = versioncmp($::kubernetes::version, '1.11.0') >= 0

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

  if $post_1_11 {
    file{$config_file:
      ensure  => file,
      mode    => '0640',
      owner   => 'root',
      group   => $kubernetes::group,
      content => template('kubernetes/kube-proxy-config.yaml.erb'),
      notify  => Service["${service_name}.service"],
    }
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
    ensure  => $service_ensure,
    enable  => true,
    require => Package[$conntrack_package_name],
  }
}
