# class kubernetes::kubectl
class kubernetes::kubectl(
  $ca_file = undef,
  $cert_file = undef,
  $key_file = undef,
){
  require ::kubernetes

  $namespace = 'kube-system'

  $kubeconfig_path = "${::kubernetes::config_dir}/kubeconfig-kubectl"
  file{$kubeconfig_path:
    ensure  => file,
    mode    => '0640',
    owner   => 'root',
    group   => $kubernetes::group,
    content => template('kubernetes/kubeconfig.erb'),
  }

  kubernetes::symlink{'kubectl':}
}
