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
  $service_kube_addon_manager = 'kube-addon-manager.service'
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

  if $kubernetes::_apiserver_insecure_port == 0 {
    $server_port = $kubernetes::apiserver_secure_port
    $protocol = 'https'
  } else {
    $server_port = $kubernetes::_apiserver_insecure_port
    $protocol = 'http'
  }

  $command = "/bin/bash -c \"while true; do if [[ \$(curl -k -w '%{http_code}' -s -o /dev/null ${protocol}://localhost:${server_port}/healthz) == 200 ]]; then break; fi; done; kubectl apply -f '${apply_file}' || rm -f '${apply_file})'\""

  # validate file first
  exec{"validate_${name}":
      path        => [
        $::kubernetes::_dest_dir,
        '/usr/bin',
        '/bin',
      ],
      environment => [
        "KUBECONFIG=${::kubernetes::kubectl::kubeconfig_path}",
      ],
      refreshonly => true,
      command     => $command,
      require     => [ Service[$service_apiserver], Service[$service_kube_addon_manager] ],
      logoutput   => true,
  }
}
