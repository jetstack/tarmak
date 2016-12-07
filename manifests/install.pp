define etcd::install (
  String $etcd_version,
)
{
  wget::fetch { "download etcd version ${etcd_version}":
    source      => "https://github.com/coreos/etcd/releases/download/v${etcd_version}/etcd-v${etcd_version}-linux-amd64.tar.gz",
    destination => '/root/',
    before      => Exec["untar etcd version ${etcd_version}"],
  }

  exec { "untar etcd version ${etcd_version}":
    command => "/bin/tar -xvzf /root/etcd-v${etcd_version}-linux-amd64.tar.gz -C /root/",
    creates => "/root/etcd-v${etcd_version}-linux-amd64/etcd",
  }

  file { "install etcd version ${etcd_version}":
    path    => "/bin/etcd-${etcd_version}",
    source  => "/root/etcd-v${etcd_version}-linux-amd64/etcd",
    mode    => '0755',
    require => Exec["untar etcd version ${etcd_version}"],
  }

  file { "install etcdctl version ${etcd_version}":
    path    => "/bin/etcdctl-${etcd_version}",
    source  => "/root/etcd-v${etcd_version}-linux-amd64/etcdctl",
    mode    => '0755',
    require => Exec["untar etcd version ${etcd_version}"],
  }
}
