class tarmak::vault (
  String $data_dir = '/var/lib/consul',
  String $dest_dir = '/opt/bin',
  String $systemd_dir = '/etc/systemd/system',
){

  include ::vault_server

  if $vault_server::cloud_provider == 'aws' {
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
      volume_id       => $::vault_server::volume_id,
      device          => "/dev/${ebs_device}",
      dest_path       => $data_dir,
      is_not_attached => $is_not_attached,
    }
  }
}
