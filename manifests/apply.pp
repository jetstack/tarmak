# adds resources to a kubernetes master
define kubernetes::apply(
  $manifests = [],
  $force = false,
  $format = 'yaml',
  $systemd_wants = [],
  $systemd_requires = [],
  $systemd_after = [],
  $systemd_before = [],
  Enum['manifests','concat'] $type = 'manifests',
){
  require ::kubernetes
  require ::kubernetes::kubectl

  if ! defined(Class['kubernetes::apiserver']) {
    fail('This defined type can only be used on the kubernetes master')
  }

  $service_apiserver = 'kube-apiserver.service'

  $_systemd_wants = $systemd_wants
  $_systemd_requires = [$service_apiserver] + $systemd_requires
  $_systemd_after = ['network.target', $service_apiserver] + $systemd_after
  $_systemd_before = $systemd_before

  $service_name = "kubectl-apply-${name}"
  $manifests_content = $manifests.join("\n---\n")
  $apply_file = "${::kubernetes::apply_dir}/${name}.${format}"
  $kubectl_path = "${::kubernetes::bin_dir}/kubectl"
  $curl_path = $::kubernetes::curl_path

  case $type {
    'manifests': {
      file{$apply_file:
        ensure  => file,
        mode    => '0640',
        owner   => 'root',
        group   => $kubernetes::group,
        content => $manifests_content,
        notify  => Service["${service_name}.service"],
      }
    }
    'concat': {
      concat { $apply_file:
        ensure  => present,
        mode    => '0640',
        owner   => 'root',
        group   => $kubernetes::group,
        notify  => Service["${service_name}.service"],
      }
    }
    default: {
      fail("Unknown type parameter: '${type}'")
    }
  }

  file{"${::kubernetes::systemd_dir}/${service_name}.service":
    ensure  => file,
    mode    => '0644',
    owner   => 'root',
    group   => 'root',
    content => template('kubernetes/kubectl-apply.service.erb'),
    notify  => [
      Service["${service_name}.service"],
    ]
  } ~>
  exec { "${service_name}-daemon-reload":
    command     => 'systemctl daemon-reload',
    path        => $::kubernetes::path,
    refreshonly => true,
  } ->
  service{ "${service_name}.service":
    ensure  => 'running',
    enable  => true,
    require => [
      Service[$service_apiserver],
    ]
  }
}
