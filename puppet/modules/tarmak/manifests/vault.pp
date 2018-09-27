class tarmak::vault (
  String $volume_id = '',
  String $data_dir = '/var/lib/consul',
  String $dest_dir = '/opt/bin',
  String $systemd_dir = '/etc/systemd/system',
  Enum['aws', ''] $cloud_provider = '',
){

  if $cloud_provider == '' and defined('$::cloud_provider') {
    $_cloud_provider = $::cloud_provider
  } else {
    $_cloud_provider = $cloud_provider
  }

  if $_cloud_provider == 'aws' {
    $disks = aws_ebs::disks()
    case $disks.length {
      0: {$ebs_device = ''; $is_not_attached = true}
      1: {$ebs_device = 'xvdd'; $is_not_attached = true}
      default: {$ebs_device = $disks[1]; $is_not_attached = false}
    }

    class{'::aws_ebs':
      bin_dir     => $dest_dir,
      systemd_dir => $systemd_dir,
    }
    aws_ebs::mount{'vault':
      volume_id     => $volume_id,
      device        => "/dev/${ebs_device}",
      dest_path     => $data_dir,
      is_not_attached => $is_not_attached,
    }
  }
}
