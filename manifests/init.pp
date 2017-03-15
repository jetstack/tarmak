class aws_ebs(
  String $systemd_dir = '/etc/systemd/system',
  String $bin_dir = '/opt/bin',
  String $dest_dir = '/opt',
){

  $path = defined('$::path') ? {
    default => '/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/opt/bin',
    true    => $::path,
  }

  ensure_resource('package', ['curl', 'gawk', 'util-linux', 'awscli', 'xfsprogs'],{
    ensure => present
  })

  $attach_bin_path = "${bin_dir}/aws_ebs_attach_volume.sh"
  file { $attach_bin_path:
    ensure  => file,
    content => template('aws_ebs/attach_volume.sh.erb'),
    mode    => '0755',
  }

  $format_bin_path = "${bin_dir}/aws_ebs_ensure_volume_formatted.sh"
  file {$format_bin_path:
    ensure  => file,
    content => template('aws_ebs/ensure_volume_formatted.sh.erb'),
    mode    => '0755',
  }
}
