# class kubernetes::kubelet

# @param cgroup_kubernetes_name name of cgroup slice for kubernetes related processes
# @param cgroup_kubernetes_reserved_memory memory reserved for kubernetes related processes
# @param cgroup_kubernetes_reserved_cpu CPU reserved for kubernetes related processes
# @param cgroup_system_name name of cgroup slice for system processes
# @param cgroup_system_reserved_memory memory reserved for system processes
# @param cgroup_system_reserved_cpu CPU reserved for system processes
class kubernetes::kubelet(
  String $role = 'worker',
  String $container_runtime = 'docker',
  String $kubelet_dir = '/var/lib/kubelet',
  String $hard_eviction_memory_threshold =
    five_percent_of_total_ram(dig44($facts, ['memory', 'system', 'total_bytes'], 1)),
  Optional[String] $network_plugin = undef,
  Integer $network_plugin_mtu = 1460,
  Boolean $allow_privileged = true,
  Boolean $register_node = true,
  Optional[Boolean] $register_schedulable = undef,
  Optional[String] $ca_file = undef,
  Optional[String] $cert_file = undef,
  Optional[String] $key_file = undef,
  Optional[String] $client_ca_file = undef,
  $node_labels = undef,
  $node_taints = undef,
  $pod_cidr = undef,
  $hostname_override = undef,
  Enum['systemd', 'cgroupfs'] $cgroup_driver =  $::osfamily ? {
    'RedHat' => 'systemd',
    default  => 'cgroupfs',
  },
  String $cgroup_root = '/',
  Optional[String] $cgroup_kube_name = '/podruntime.slice',
  Optional[String] $cgroup_kube_reserved_memory = '256Mi',
  Optional[String] $cgroup_kube_reserved_cpu = '10m',
  Optional[String] $cgroup_system_name = '/system.slice',
  Optional[String] $cgroup_system_reserved_memory = '128Mi',
  Optional[String] $cgroup_system_reserved_cpu = '10m',
  Array[String] $systemd_wants = [],
  Array[String] $systemd_requires = [],
  Array[String] $systemd_after = [],
  Array[String] $systemd_before = [],
){
  require ::kubernetes

  $_systemd_wants = $systemd_wants
  if $container_runtime == 'docker' {
    $_systemd_after = ['docker.service'] + $systemd_after
    $_systemd_requires = ['docker.service'] + $systemd_after
  } else {
    $_systemd_after = $systemd_after
    $_systemd_requires = $systemd_after
  }
  $_systemd_before = $systemd_before

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

  if $node_taints == undef {
    if !$_register_schedulable {
      $_node_taints = {
        'node-role.kubernetes.io/master' => ':NoSchedule',
      }
    }
    else {
      $_node_taints = {}
    }
  } else {
    $_node_taints = $node_taints
  }

  $_node_taints_list = $_node_taints.map |$k,$v| { "${k}=${v}"}
  $_node_taints_string = $_node_taints_list.join(',')

  if $node_labels == undef {
    $_node_labels = {
      'role'                            => $role,
      "node-role.kubernetes.io/${role}" => '',
    }
  } else {
    $_node_labels = $node_labels
  }

  $_node_labels_list = $_node_labels.map |$k,$v| { "${k}=${v}"}
  $_node_labels_string = $_node_labels_list.join(',')

  $cluster_domain = $::kubernetes::cluster_domain
  $cluster_dns = $::kubernetes::_cluster_dns
  $cloud_provider = $::kubernetes::cloud_provider

  # TODO: this should come from something higher in the stack
  $container_interface = 'cali+'

  $service_name = 'kubelet'

  if $ca_file == undef {
    $_ca_file = '/var/run/kubernetes/apiserver.crt'
  } else {
    $_ca_file = $ca_file
  }

  $seltype = 'container_file_t'
  $seltype_old = 'svirt_sandbox_file_t'
  file{$kubelet_dir:
    ensure  => 'directory',
    mode    => '0750',
    owner   => 'root',
    group   => 'root',
    seltype => $seltype,
  }

  if dig44($facts, ['os', 'selinux', 'enabled'], false) and $::osfamily == "Redhat" {
    $policy_package = 'selinux-policy-targeted'
    ensure_resource('package', $policy_package, {
      ensure => 'latest',
    })

    exec { 'semanage_fcontext_kubelet_dir':
      # fall back to old seltype, if command fails
      command => "semanage fcontext -a -t ${seltype} \"${kubelet_dir}(/.*)?\" || semanage fcontext -a -t ${seltype_old} \"${kubelet_dir}(/.*)?\"",
      unless  => "semanage fcontext -l | grep \"${kubelet_dir}(/.*).*:${seltype}\"",
      require => File[$kubelet_dir],
      path    => $::kubernetes::path
    }

    file{"${kubelet_dir}/pods":
      ensure  => 'directory',
      mode    => '0750',
      owner   => 'root',
      group   => 'root',
      seltype => $seltype,
      require => [File[$kubelet_dir], Package[$policy_package]],
    }

    file{"${kubelet_dir}/plugins":
      ensure  => 'directory',
      mode    => '0750',
      owner   => 'root',
      group   => 'root',
      seltype => $seltype,
      require => [File[$kubelet_dir], Package[$policy_package]],
    }

    $availability_zone = dig44($facts, ['ec2_metadata', 'placement', 'availability-zone'])
    if $cloud_provider == 'aws' and $availability_zone != undef {
      file{"${kubelet_dir}/plugins/kubernetes.io":
        ensure  => 'directory',
        mode    => '0750',
        owner   => 'root',
        group   => 'root',
        seltype => $seltype,
        require => File["${kubelet_dir}/plugins"],
      }
      -> file{"${kubelet_dir}/plugins/kubernetes.io/aws-ebs":
        ensure  => 'directory',
        mode    => '0750',
        owner   => 'root',
        group   => 'root',
        seltype => $seltype,
      }
      -> file{"${kubelet_dir}/plugins/kubernetes.io/aws-ebs/mounts":
        ensure  => 'directory',
        mode    => '0750',
        owner   => 'root',
        group   => 'root',
        seltype => $seltype,
      }
      -> file{"${kubelet_dir}/plugins/kubernetes.io/aws-ebs/mounts/aws":
        ensure  => 'directory',
        mode    => '0750',
        owner   => 'root',
        group   => 'root',
        seltype => $seltype,
      }
      -> file{"${kubelet_dir}/plugins/kubernetes.io/aws-ebs/mounts/aws/${availability_zone}":
        ensure  => 'directory',
        mode    => '0750',
        owner   => 'root',
        group   => 'root',
        seltype => $seltype,
      }
    }
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

  kubernetes::symlink{'kubelet':}
  -> file{"${::kubernetes::systemd_dir}/${service_name}.service":
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

  # ensure socat installed (for portforward)
  ensure_resource('package', 'socat',{
    ensure => 'present',
  })

}
