# deletes resources to a kubernetes master
define kubernetes::delete(
  $format = 'yaml',
){
  require ::kubernetes
  require ::kubernetes::kubectl

  $service_apiserver = 'kube-apiserver.service'

  $apply_file = "${::kubernetes::apply_dir}/${name}.${format}"
  if $kubernetes::_apiserver_insecure_port == 0 {
    $server_port = $kubernetes::apiserver_secure_port
    $protocol = 'https'
  } else {
    $server_port = $kubernetes::_apiserver_insecure_port
    $protocol = 'http'
  }

  $command = "/bin/bash -c \"while true; do if [[ \$(curl -k -w '%{http_code}' -s -o /dev/null ${protocol}://localhost:${server_port}/healthz) == 200 ]]; then break; else sleep 2; fi; done; kubectl delete -f '${apply_file}'; rm -f '${apply_file}'; exit 0\""

  exec{"delete_${name}":
    path        => [
      $::kubernetes::_dest_dir,
      '/usr/bin',
      '/bin',
    ],
    environment => [
      "KUBECONFIG=${::kubernetes::kubectl::kubeconfig_path}",
    ],
    command     => $command,
    onlyif      => "test -e ${apply_file}",
    require     => [ Service[$service_apiserver] ],
  }
}
