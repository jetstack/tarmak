# This module attaches, formats (if needed) and mounts EBS volumes in AWS. This
# base class just makes sure that all the necessary dependencies are met. To
# actually attach & mount a volume you have to use the defined type
# `aws_ebs::mount`
#
# @example Declaring the base class
#   include ::aws_ebs
# @example Override binary directory (needs to exist)
#   class{'aws_ebs':
#     bin_dir => '/usr/local/sbin',
#   }
#
# @param bin_dir path to the binary directory for helper scripts
# @param systemd_dir path to the directory where systemd units should be placed
class aws_ebs(
  String $systemd_dir = '/etc/systemd/system',
  String $bin_dir = '/opt/bin',
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
