class site_module::docker_config {
  file { '/etc/sysconfig/docker':
    ensure  => file,
    content => template('site_module/docker.erb'),
  }

  if $kubernetes::kubelet::cgroup_kube_name {

    $cgroup_kube_basename = regsubst( $kubernetes::kubelet::cgroup_kube_name, '^\/', '')

    file { '/etc/systemd/system/docker.service.d':
      ensure  => directory,
    } -> file { '/etc/systemd/system/docker.service.d/10-slice.conf':
      ensure  => directory,
      content => "[Service]\nSlice=${cgroup_kube_basename}\n",
    }

  }
}
