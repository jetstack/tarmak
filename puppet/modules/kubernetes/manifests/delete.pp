# adds resources to a kubernetes master
define kubernetes::delete(
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

  $service_name_apply = "kubectl-apply-${name}"
  $service_name_delete = "kubectl-delete-${name}"
  $manifests_content = $manifests.join("\n---\n")
  $apply_file = "${::kubernetes::apply_dir}/${name}.${format}"
  $delete_file = "${::kubernetes::delete_dir}/${name}.${format}"
  $kubectl_path = "${::kubernetes::bin_dir}/kubectl"
  $curl_path = $::kubernetes::curl_path

  exec {"check_${apply_file}_exist":
    command => '/bin/true',
    onlyif  => "/bin/test -f ${::kubernetes::systemd_dir}/${service_name_apply}.service"
  }

  case $type {
    'manifests': {
      file{$delete_file:
        ensure  => file,
        mode    => '0640',
        owner   => 'root',
        group   => $kubernetes::group,
        content => $manifests_content,
        notify  => Service["${service_name_delete}.service"],
        require => Exec["check_${apply_file}_exist"],
      }
    }
    'concat': {
      concat { $delete_file:
        ensure         => absent,
        ensure_newline => true,
        mode           => '0640',
        owner          => 'root',
        group          => $kubernetes::group,
        notify         => Service["${service_name_delete}.service"],
        require        => Exec["check_${apply_file}_exist"],
      }
    }
    default: {
      fail("Unknown type parameter: '${type}'")
    }
  }

  case $type {
    'manifests': {
      file{$apply_file:
        ensure => absent,
      }
    }
    'concat': {
      concat { $apply_file:
        ensure => absent,
      }
    }
    default: {
      fail("Unknown type parameter: '${type}'")
    }
  }

  file{"${::kubernetes::systemd_dir}/${service_name_delete}.service":
    ensure  => file,
    mode    => '0644',
    owner   => 'root',
    group   => 'root',
    content => template('kubernetes/kubectl-delete.service.erb'),
    notify  => [
      Service["${service_name_delete}.service"],
    ],
    require => Exec["check_${apply_file}_exist"],
  }
  ~> exec { "${service_name_delete}-daemon-reload":
    command     => 'systemctl daemon-reload',
    path        => $::kubernetes::path,
    refreshonly => true,
  }
  -> service{ "${service_name_delete}.service":
    ensure  => 'running',
    enable  => true,
    require => [
      Service[$service_apiserver],
    ]
  }
  -> file{"${::kubernetes::systemd_dir}/${service_name_apply}.service":
    ensure => absent,
  }
}
