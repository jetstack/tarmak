# class kubernetes::master
class kubernetes::addon_manager(
  $systemd_wants = [],
  $systemd_requires = [],
  $systemd_after = [],
  $systemd_before = [],
)  {
  require ::kubernetes
  require ::kubernetes::kubectl

  if ! defined(Class['kubernetes::apiserver']) {
    fail('This defined type can only be used on the kubernetes master')
  }

  $service_apiserver = 'kube-apiserver.service'
  $service_controller_manager = 'kube-controller-manager.service'

  # TOOD: add kubeconfig/kubectl service
  $_systemd_wants = $systemd_wants
  $_systemd_requires = [$service_apiserver, $service_controller_manager] + $systemd_requires
  $_systemd_after = ['network.target', $service_apiserver, $service_controller_manager] + $systemd_after
  $_systemd_before = $systemd_before

  $service_name = 'kube-addon-manager'

  $kubeconfig_path = $::kubernetes::kubectl::kubeconfig_path
  $manifest_dir = $::kubernetes::apply_dir
  $kubectl_path = "${::kubernetes::_dest_dir}/kubectl"
  $kube_addon_manager_path = "${::kubernetes::_dest_dir}/kube-addon-manager"

  file {$kube_addon_manager_path:
    ensure  => file,
    mode    => '0755',
    owner   => 'root',
    group   => $kubernetes::group,
    content => template('kubernetes/kube-addon-manager.sh.erb'),
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
