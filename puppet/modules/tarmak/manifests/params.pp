# Defines parameters for other classes to reuse
# @private
class tarmak::params{
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
  $kubernetes_version = '1.10.6'

  ## etcd
  $etcd_advertise_client_network = '172.16.0.0/12'

  ## fluent-bit
  $fluent_bit_configs = []

}
