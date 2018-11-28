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
  require ::kubernetes::addon_manager

  if ! defined(Class['kubernetes::apiserver']) {
    fail('This defined type can only be used on the kubernetes master')
  }

  $service_apiserver = 'kube-apiserver.service'
  $manifests_content = $manifests.join("\n---\n")
  $apply_file = "${::kubernetes::apply_dir}/${name}.${format}"

  case $type {
    'manifests': {
      file{$apply_file:
        ensure  => file,
        mode    => '0640',
        owner   => 'root',
        group   => $kubernetes::group,
        content => $manifests_content,
        notify  => Exec["validate_${name}"],
      }
    }
    'concat': {
      concat { $apply_file:
        ensure         => present,
        ensure_newline => true,
        mode           => '0640',
        owner          => 'root',
        group          => $kubernetes::group,
        notify         => Exec["validate_${name}"],
      }
    }
    default: {
      fail("Unknown type parameter: '${type}'")
    }
  }

  # validate file first
  exec{"validate_${name}":
      path        => [
        $::kubernetes::_dest_dir,
        '/usr/bin',
        '/bin',
      ],
      refreshonly => true,
      command     => "kubectl apply -f '${apply_file}' || rm -f '${apply_file}'",
      require     => Service[$service_apiserver],
  }
}
