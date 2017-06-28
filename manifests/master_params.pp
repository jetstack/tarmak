# == Class kubernetes::params
class kubernetes::master_params {
  $cluster_dns = 'cluster.local'
  $cluster_ip = '10.254.0.0'
  $cluster_ip_mask = 16
  $allow_privileged = false
}
