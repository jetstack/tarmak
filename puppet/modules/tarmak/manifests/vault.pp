class tarmak::vault (
  String $data_dir = '/var/lib/consul',
  String $dest_dir = '/opt/bin',
  String $systemd_dir = '/etc/systemd/system',
  Enum['aws', ''] $cloud_provider = '',
){
  require ::vault_server

  if $cloud_provider == 'aws' {
    $disks = aws_ebs::disks()
    case $disks.length {
      0: {$ebs_device = ''}
      1: {$ebs_device = $disks[0]}
      default: {$ebs_device = $disks[1]}
    }

    class{'::aws_ebs':
      bin_dir     => $dest_dir,
      systemd_dir => $systemd_dir,
    }
    aws_ebs::mount{'vault':
      volume_id => $vault_server::volume_id,
      device    => $ebs_device,
      dest_path => $data_dir,
    }
  }
}
