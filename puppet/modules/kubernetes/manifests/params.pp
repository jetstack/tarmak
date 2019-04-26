# == Class kubernetes::params
class kubernetes::params {
  $version = '1.10.6'
  $bin_dir = '/opt/bin'
  $dest_dir = '/opt'
  $config_dir = '/etc/kubernetes'
  $run_dir = '/var/run/kubernetes'
  $apply_dir = '/etc/kubernetes/apply'
  $download_dir = '/tmp'
  $systemd_dir = '/etc/systemd/system'
  $download_url = 'https://storage.googleapis.com/kubernetes-release/release/v#VERSION#/bin/linux/amd64/hyperkube'
  $aws_authenticator_download_url = 'https://github.com/kubernetes-sigs/aws-iam-authenticator/releases/download/v#VERSION#/aws-iam-authenticator_#VERSION#_linux_amd64'
  $aws_authenticator_version = '0.4.0'
  $sysctl_dir = '/etc/sysctl.d'
  $log_level = '1'
  $uid = 873
  $gid = 873
  $user = 'kubernetes'
  $group = 'kubernetes'
  $curl_path = $::osfamily ? {
    'RedHat' => '/bin/curl',
    'Debian' => '/usr/bin/curl',
    default  => '/usr/bin/curl',
  }
}
