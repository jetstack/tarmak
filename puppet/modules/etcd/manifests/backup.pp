# Backup an instance of an etcd server
define etcd::backup (
  String $version,
  Integer $client_port,
  String $service_name,
  Boolean $tls = false,
  String $local_backup_dir = '/tmp/etcd-backup',
  String $tls_cert_path = nil,
  String $tls_key_path = nil,
  String $tls_ca_path = nil,
  Optional[Enum['aws:kms','']] $sse = undef,
  Optional[String] $bucket_prefix = undef,
  Enum['file', 'absent'] $file_ensure = 'file',
  Enum['running', 'stopped'] $service_ensure = 'running',
  Boolean $service_enable = true,
){
  include ::etcd

  $hostname = $::hostname
  $backup_service_name = "${service_name}-backup"
  $backup_script_path = "${::etcd::params::bin_dir}/etcd-${name}-backup.sh"
  $etcdctl_path = "${::etcd::dest_dir}/${::etcd::params::app_name}-${version}/etcdctl"

  $config_dir = $::etcd::config_dir

  $_bucket_prefix = pick_default($bucket_prefix, $::etcd::backup_bucket_prefix)
  if $_bucket_prefix == undef or $_bucket_prefix == '' {
    fail('no backup_bucket_prefix set')
  }

  $_sse = pick_default($sse, $::etcd::backup_sse, '')
  $_bucket_endpoint = $::etcd::backup_bucket_endpoint

  $aws_s3_args = $_sse ? {
    'aws:kms' => ['--sse','aws:kms'],
    default   => [],
  } + $_bucket_endpoint ? {
    ''      => [],
    default =>  ['--endpoint', $_bucket_endpoint],
  }

  if $tls {
    $proto = 'https'
  } else {
    $proto = 'http'
  }
  $endpoints = "${proto}://127.0.0.1:${client_port}"

  $hour = fqdn_rand(24, $name)
  $backup_schedule = "*-*-* ${hour}:00:00"

  ensure_resource('file', [$::etcd::params::bin_dir], {
    ensure => directory,
    mode   => '0755',
  })

  File[$::etcd::params::bin_dir]
  -> file { $backup_script_path:
    ensure  => file,
    content => template('etcd/etcd-backup.sh.erb'),
    mode    => '0755'
  }

  ensure_resource('package', ['awscli'],{
    ensure => present
  })


  file { "${etcd::systemd_dir}/${backup_service_name}.service":
    ensure  => $file_ensure,
    content => template('etcd/etcd-backup.service.erb'),
    notify  => Exec["${name}-systemctl-daemon-reload"],
    mode    => '0644'
  }

  file { "${etcd::systemd_dir}/${backup_service_name}.timer":
    ensure  => $file_ensure,
    content => template('etcd/etcd-backup.timer.erb'),
    notify  => Exec["${name}-systemctl-daemon-reload"],
  }
  ~> service { "${backup_service_name}.timer":
    ensure  => $service_ensure,
    enable  => $service_enable,
    require => [
      Exec["${name}-systemctl-daemon-reload"],
      Package['awscli'],
      File[$backup_script_path],
    ],
  }
}
