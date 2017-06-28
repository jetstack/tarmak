# This defined type attaches, formats (if needed) and mounts a single EBS
# volume in AWS.
#
# @example Attach, format & mount EBS volume
#   aws_ebs::mount{'data':
#     volume_id => 'vol-deadbeef',
#     device    => '/dev/xvdd',
#     dest_path => '/mnt',
#   }
#
# @param volume_id the volume id of the AWS EBS volume
# @param dest_path where to mount the device (needs to exists)
# @param device block device to attach to (should be `/dev/xvd[a-z]`)
# @param filesystem select the filesystem to initialize a volume
define aws_ebs::mount(
  String $volume_id,
  String $dest_path,
  String $device,
  Enum['xfs'] $filesystem = 'xfs',
){
  require ::aws_ebs

  $systemd_reload = "aws-ebs-mount-${name}-systemctl-daemon-reload"
  exec { $systemd_reload:
    command     => 'systemctl daemon-reload',
    refreshonly => true,
    path        => $::aws_ebs::path,
  }

  $attach_service_name = "attach-ebs-volume-${name}.service"
  $format_service_name = "ensure-ebs-volume-${name}-formatted.service"
  $mount_name = regsubst(regsubst($dest_path, '/', '-', 'G'), '^.(.*)$', '\1')
  $mount_service_name = "${mount_name}.mount"

  file { "${::aws_ebs::systemd_dir}/${attach_service_name}":
    ensure  => file,
    mode    => '0644',
    content => template('aws_ebs/attach-volume.service.erb'),
    notify  => Exec[$systemd_reload],
  } ~>
  service { $attach_service_name:
    ensure  => running,
    enable  => true,
    before  => Service[$format_service_name],
    require => Exec[$systemd_reload],
  }

  file { "${::aws_ebs::systemd_dir}/${format_service_name}":
    ensure  => file,
    mode    => '0644',
    content => template('aws_ebs/ensure-volume-formatted.service.erb'),
    notify  => Exec[$systemd_reload],
  } ~>
  service { $format_service_name:
    ensure  => running,
    enable  => true,
    before  => Service[$mount_service_name],
    require => Exec[$systemd_reload],
  }

  file { "${::aws_ebs::systemd_dir}/${mount_service_name}":
    ensure  => file,
    mode    => '0644',
    content => template('aws_ebs/volume.mount.erb'),
    notify  => Exec[$systemd_reload],
  } ~>
  service { $mount_service_name:
    ensure  => running,
    require => Exec[$systemd_reload],
  }
}
