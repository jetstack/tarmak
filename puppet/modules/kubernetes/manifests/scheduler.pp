# class kubernetes::master
class kubernetes::scheduler(
  $ca_file = undef,
  $cert_file = undef,
  $key_file = undef,
  $systemd_wants = [],
  $systemd_requires = [],
  $systemd_after = [],
  $systemd_before = [],
  Hash[String,Boolean] $feature_gates = {},
)  {
  require ::kubernetes

  $_systemd_wants = $systemd_wants
  $_systemd_requires = $systemd_requires
  $_systemd_after = ['network.target'] + $systemd_after
  $_systemd_before = $systemd_before

  $post_1_15 = versioncmp($::kubernetes::version, '1.15.0') >= 0

  $service_name = 'kube-scheduler'

  if $post_1_15 {
    $command_name = 'kube-scheduler'
  } else {
    $command_name = 'scheduler'
  }

  $kubeconfig_path = "${::kubernetes::config_dir}/kubeconfig-scheduler"

  if $feature_gates == {} {
    $_feature_gates = $::kubernetes::_enable_pod_priority ? {
      true    => {'PodPriority' => true},
      default => undef,
    }
  } else {
    $_feature_gates = $feature_gates
  }

  kubernetes::symlink{$command_name:}
  -> file{$kubeconfig_path:
    ensure  => file,
    mode    => '0640',
    owner   => 'root',
    group   => $kubernetes::group,
    content => template('kubernetes/kubeconfig.erb'),
    notify  => Service["${service_name}.service"],
  }

  file{"${::kubernetes::systemd_dir}/${service_name}.service":
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
    ensure => running,
    enable => true,
  }
}
