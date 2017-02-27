# class kubernetes::kubelet
class kubernetes::kubelet(
  $role = 'worker',
  $container_runtime = 'docker',
  $kubelet_dir = '/var/lib/kubelet',
  $network_plugin = undef,
  $network_plugin_mtu = 1460,
  $allow_privileged = true,
  $register_node = true,
  $register_schedulable = undef,
  $ca_file = undef,
  $cert_file = undef,
  $key_file = undef,
  $node_labels = undef,
  $pod_cidr = undef,
){
  require ::kubernetes

  if $register_schedulable == undef {
    if $role == 'master' {
      $_register_schedulable = false
    }
    else {
      $_register_schedulable = true
    }
  } else {
    $_register_schedulable = $register_schedulable
  }

  if $node_labels == undef {
    $_node_labels = {
      'role' => $role,
    }
  } else {
    $_node_labels = $node_labels
  }

  $_node_labels_list = $_node_labels.map |$k,$v| { "${k}=${v}"}
  $_node_labels_string = $_node_labels_list.join(',')

  $cluster_domain = $::kubernetes::cluster_domain
  $cluster_dns = $::kubernetes::_cluster_dns

  $service_name = 'kubelet'

  if $ca_file == undef {
    $_ca_file = '/var/run/kubernetes/apiserver.crt'
  } else {
    $_ca_file = $ca_file
  }


  $seltype = 'svirt_sandbox_file_t'
  file{$kubelet_dir:
    ensure  => 'directory',
    mode    => '0750',
    owner   => 'root',
    group   => 'root',
    seltype => $seltype,
  }

  if dig($facts, ['os','selinux','enabled'], false) {
    exec { 'semanage_fcontext_kubelet_dir':
      command => "semanage fcontext -a -t ${seltype} \"${kubelet_dir}(/.*)?\"",
      unless  => "semanage fcontext -l | grep \"${kubelet_dir}(/.*).*:${seltype}\"",
      require => File[$kubelet_dir],
      path    => $::kubernetes::path
    }
  }

  file{"${kubelet_dir}/pods":
    ensure  => 'directory',
    mode    => '0750',
    owner   => 'root',
    group   => 'root',
    seltype => $seltype,
    require => File[$kubelet_dir],
  }

  file{"${kubelet_dir}/plugins":
    ensure  => 'directory',
    mode    => '0750',
    owner   => 'root',
    group   => 'root',
    seltype => $seltype,
    require => File[$kubelet_dir],
  }

  $kubeconfig_path = "${::kubernetes::config_dir}/kubeconfig-kubelet"
  file{$kubeconfig_path:
    ensure  => file,
    mode    => '0640',
    owner   => 'root',
    group   => $kubernetes::group,
    content => template('kubernetes/kubeconfig.erb'),
    notify  => Service["${service_name}.service"],
  }

  kubernetes::symlink{'kubelet':} ->
  file{"${::kubernetes::systemd_dir}/${service_name}.service":
    ensure  => file,
    mode    => '0640',
    owner   => 'root',
    group   => 'root',
    content => template("kubernetes/${service_name}.service.erb"),
    notify  => Service["${service_name}.service"],
  } ~>
  exec { "${service_name}-daemon-reload":
    command     => 'systemctl daemon-reload',
    path        => $::kubernetes::path,
    refreshonly => true,
  } ->
  service{ "${service_name}.service":
    ensure => running,
    enable => true,
  }

  # ensure socat installed (for portforward)
  ensure_resource('package', 'socat',{
    ensure => 'present',
  })

}
