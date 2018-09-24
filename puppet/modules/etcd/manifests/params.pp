# etcd variable defaults
class etcd::params {
  $app_name = 'etcd'
  $user = 'etcd'
  $group = 'etcd'
  $uid = 873
  $gid = 873
  $version = '3.2.24'
  $bin_dir = '/opt/bin'
  $dest_dir = '/opt'
  $config_dir = '/etc/etcd'
  $data_dir = '/var/lib/etcd'
  $download_dir = '/tmp'
  $download_url = 'https://storage.googleapis.com/etcd/v#VERSION#/etcd-v#VERSION#-linux-amd64.tar.gz'
  $systemd_dir = '/etc/systemd/system'
}
