class puppernetes::params{
  $cluster_name = 'cluster'
  $dns_root = 'jetstack.net'
  $hostname = $::hostname

  ## General
  $helper_path = '/usr/local/sbin'
  $systemctl_path = $::osfamily ? {
    'RedHat' => '/usr/bin/systemctl',
    'Debian' => '/bin/systemctl',
    default  => '/usr/bin/systemctl',
  }

  ## Kubernetes
  $kubernetes_version = '1.5.4'

  ## etcd
  $etcd_advertise_client_network = '172.16.0.0/12'

}
