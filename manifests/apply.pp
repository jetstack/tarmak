# adds resources to a kubernetes master
define kubernetes::apply(
  $manifests = [],
  $force = false,
  $format = 'yaml',
  $systemd_wants = [],
  $systemd_requires = [],
  $systemd_after = [],
  $systemd_before = [],
){
  require ::kubernetes

  if ! defined(Class['kubernetes::apiserver']) {
    fail('This defined type can only be used on the kubernetes master')
  }

  ensure_resource('kubernetes::symlink', 'kubectl',{})

  $service_apiserver = 'kube-apiserver.service'

  $_systemd_wants = $systemd_wants
  $_systemd_requires = [$service_apiserver] + $systemd_requires
  $_systemd_after = ['network.target', $service_apiserver] + $systemd_after
  $_systemd_before = $systemd_before

  $service_name = "kubectl-apply-${name}"
  $manifests_content = $manifests.join("\n---\n")
  $apply_file = "${::kubernetes::apply_dir}/${name}.${format}"
  $kubectl_path = "${::kubernetes::bin_dir}/kubectl"
  $curl_path = '/bin/curl'


  file{$apply_file:
    ensure  => file,
    mode    => '0640',
    owner   => 'root',
    group   => $kubernetes::group,
    content => $manifests_content,
    notify  => Exec["${service_name}-trigger"],
  }

  file{"${::kubernetes::systemd_dir}/${service_name}.service":
    ensure  => file,
    mode    => '0644',
    owner   => 'root',
    group   => 'root',
    content => template('kubernetes/kubectl-apply.service.erb'),
    notify  => [
      Service["${service_name}.service"],
      Exec["${service_name}-trigger"],
    ]
  } ~>
  exec { "${service_name}-daemon-reload":
    command     => 'systemctl daemon-reload',
    path        => $::kubernetes::path,
    refreshonly => true,
  } ->
  service{ "${service_name}.service":
    enable  => true,
    require => Kubernetes::Symlink['kubectl'],
  } ->
  exec { "${service_name}-trigger":
    command     => "systemctl start ${service_name}.service",
    path        => $::kubernetes::path,
    refreshonly => true,
  }
}
