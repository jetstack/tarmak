class etcd_mount
{
  include ::systemd

  file { "/usr/lib/systemd/system/attach-ebs-volume.service":
    ensure => file,
    source => "puppet:///etcd_mount_module/attach-ebs-volume.service",
    before => Service["attach-ebs-volume.service"],
  } ~>
  Exec['systemctl-daemon-reload']

  file { "/usr/lib/systemd/system/format-ebs-volume.service":
    ensure => file,
    source => "puppet:///etcd-mount-module/format-ebs-volume.service",
    before => Service["format-ebs-volume.service"],
  } ~>
  Exec['systemctl-daemon-reload']

  file { "/usr/lib/systemd/system/var-lib-etcd.mount":
    ensure => file,
    source => "puppet:///etcd-mount/module/var-lib-etcd.mount",
    before => Service["var-lib-etcd.mount"],`
  } ~>
  Exec['systemctl-daemon-reload']

  file { "/usr/local/sbin/attach_volume.sh":
    ensure => file,
    source => "puppet:///etcd-mount/module/attach_volume.sh",
    before => Service["attach-ebs-volume.service"],
  }

  file { "/usr/local/sbin/format_volume.sh":
    ensure => file,
    source => "puppet:///etcd-mount/module/format_volume.sh",
    before => Service["format-ebs-volume.service"],
  }

  service { "attach-ebs-volume.service":
    enable => true,
    ensure => running,
    before => Service["format-ebs-volume.service"],
  }

  service { "format-ebs-volume.service":
    enable => true,
    ensure => running,
    before => Service["var-lib-etcd.mount"],
  }

  service { "var-lib-etcd.mount":
    enable => true,
    ensure => running,
  }
}
