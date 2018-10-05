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
  Optional[String] $eviction_hard_memory_available_threshold = '5%',
  Optional[String] $eviction_hard_nodefs_available_threshold = '10%',
  Optional[String] $eviction_hard_nodefs_inodes_free_threshold = '5%',
  Boolean $eviction_soft_enabled = true,
  Optional[String] $eviction_soft_memory_available_threshold = '10%',
  Optional[String] $eviction_soft_nodefs_available_threshold = '15%',
  Optional[String] $eviction_soft_nodefs_inodes_free_threshold = '10%',
  Optional[String] $eviction_soft_memory_available_grace_period = '0m',
  Optional[String] $eviction_soft_nodefs_available_grace_period = '0m',
  Optional[String] $eviction_soft_nodefs_inodes_free_grace_period = '0m',
  String $eviction_max_pod_grace_period = '-1',
  String $eviction_pressure_transition_period = '2m',
  Optional[String] $eviction_minimum_reclaim_memory_available = '100Mi',
  Optional[String] $eviction_minimum_reclaim_nodefs_available = '1Gi',
  Optional[String] $eviction_minimum_reclaim_nodefs_inodes_free = undef,
  Optional[String] $network_plugin = undef,
  Integer $network_plugin_mtu = 1460,
  Boolean $allow_privileged = true,
  Boolean $register_node = true,
  Optional[Boolean] $register_schedulable = undef,
  Optional[String] $ca_file = undef,
  Optional[String] $cert_file = undef,
  Optional[String] $key_file = undef,
  Optional[String] $client_ca_file = undef,
  $feature_gates = [],
  $node_labels = undef,
  $node_taints = undef,
  $pod_cidr = undef,
  $hostname_override = undef,
  Enum['systemd', 'cgroupfs'] $cgroup_driver =  $::osfamily ? {
    default  => 'cgroupfs',
  },
  String $cgroup_root = '/',
  Optional[String] $cgroup_kube_name = '/podruntime.slice',
  Optional[String] $cgroup_kube_reserved_memory = undef,
  Optional[String] $cgroup_kube_reserved_cpu = '100m',
  Optional[String] $cgroup_system_name = '/system.slice',
  Optional[String] $cgroup_system_reserved_memory = '128Mi',
  Optional[String] $cgroup_system_reserved_cpu = '100m',
  Array[String] $systemd_wants = [],
  Array[String] $systemd_requires = [],
  Array[String] $systemd_after = [],
  Array[String] $systemd_before = [],
  String $config_file = "${kubelet_dir}/kubelet-config.yaml",
){
  require ::kubernetes

  $post_1_11 = versioncmp($::kubernetes::version, '1.11.0') >= 0

  if ! $eviction_soft_memory_available_threshold or ! $eviction_soft_memory_available_grace_period {
    $_eviction_soft_memory_available_threshold = undef
    $_eviction_soft_memory_available_grace_period = undef
  } else {
    $_eviction_soft_memory_available_threshold = $eviction_soft_memory_available_threshold
    $_eviction_soft_memory_available_grace_period = $eviction_soft_memory_available_grace_period
  }

  if ! $eviction_soft_nodefs_available_threshold or ! $eviction_soft_nodefs_available_grace_period {
    $_eviction_soft_nodefs_available_threshold = undef
    $_eviction_soft_nodefs_available_grace_period = undef
  } else {
    $_eviction_soft_nodefs_available_threshold = $eviction_soft_nodefs_available_threshold
    $_eviction_soft_nodefs_available_grace_period = $eviction_soft_nodefs_available_grace_period
  }

  if ! $eviction_soft_nodefs_inodes_free_threshold or ! $eviction_soft_nodefs_inodes_free_grace_period {
    $_eviction_soft_nodefs_inodes_free_threshold = undef
    $_eviction_soft_nodefs_inodes_free_grace_period = undef
  } else {
    $_eviction_soft_nodefs_inodes_free_threshold = $eviction_soft_nodefs_inodes_free_threshold
    $_eviction_soft_nodefs_inodes_free_grace_period = $eviction_soft_nodefs_inodes_free_grace_period
  }

  if $cgroup_kube_reserved_memory == undef {
    $cgroup_kube_reserved_memory_default_mi_bytes = 1024
    $node_memory_total_mi_bytes = dig44($facts, ['memory', 'system', 'total_bytes'], 1) / ( 1024 * 1024 )
    if $node_memory_total_mi_bytes / 4 < $cgroup_kube_reserved_memory_default_mi_bytes {
      $cgroup_kube_reserved_memory_mi_bytes = $node_memory_total_mi_bytes / 4
    } else {
      $cgroup_kube_reserved_memory_mi_bytes = $cgroup_kube_reserved_memory_default_mi_bytes
    }
    $_cgroup_kube_reserved_memory = "${cgroup_kube_reserved_memory_mi_bytes}Mi"
  } else {
    $_cgroup_kube_reserved_memory = $cgroup_kube_reserved_memory
  }

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

  if $feature_gates == [] {
    $_feature_gates = delete_undef_values([
      $::kubernetes::_enable_pod_priority ? { true => 'PodPriority=true', default => undef },
    ])
  } else {
    $_feature_gates = $feature_gates
  }

  $_config_feature_gates = $_feature_gates.map |$gate| {
    $s = split($gate, '=')
    if $s.length < 2 {
      $feature = $s[0]; "${feature}: true"
    } else {
      $feature = $s[0,-2].join('='); $enable = $s[-1]; "${feature}: ${enable}"
    }
  }

  if !$_register_schedulable {
    $_default_node_taints = {
      'node-role.kubernetes.io/master' => ':NoSchedule',
    }
  } else {
    $_default_node_taints = {}
  }

  if $node_taints == undef {
    $_merged_node_taints = $_default_node_taints
  } else {
    $_merged_node_taints = $_default_node_taints+$node_taints
  }

  # format values as key=value strings, reject values set to REMOVE
  $_node_taints_string =
    delete_undef_values($_merged_node_taints.map |$k,$v| { $v ? { 'REMOVE:REMOVE' => undef, default => "${k}=${v}" } }).join(',')

  $_default_node_labels = {
    'role'                            => $role,
    "node-role.kubernetes.io/${role}" => '',
  }

  if $node_labels == undef {
    $_merged_node_labels = $_default_node_labels
  } else {
    # defaults + params, defaults overwritten using the same key
    $_merged_node_labels = $_default_node_labels+$node_labels
  }

  # format values as key=value strings, reject values set to REMOVE
  $_node_labels_string =
    delete_undef_values($_merged_node_labels.map |$k,$v| { $v ? { 'REMOVE' => undef, default => "${k}=${v}" } }).join(',')

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

  if dig44($facts, ['os', 'selinux', 'enabled'], false) and $::osfamily == 'Redhat' {
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

  if $post_1_11 {
    file{$config_file:
      ensure  => file,
      mode    => '0640',
      owner   => 'root',
      group   => $kubernetes::group,
      content => template('kubernetes/kubelet-config.yaml.erb'),
      notify  => Service["${service_name}.service"],
    }
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
