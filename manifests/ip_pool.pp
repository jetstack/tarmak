define calico::ip_pool (
  String $ip_pool,
  Integer $ip_mask,
  String $ipip_enabled
)
{
  include ::calico

  file { "/etc/calico/ipPool-${ip_pool}-${ip_mask}.yaml":
    ensure  => file,
    content => template('calico/ipPool.yaml.erb'),
    require => File["${::calico::helper_dir}/calico_helper.sh"],
  } ->
  exec { "Configure calico ipPool for CIDR ${ip_pool}-${ip_mask}":
    user    => 'root',
    command => "${::calico::helper_dir}/calico_helper.sh apply ${::calico::config_dir}/ipPool-${ip_pool}-${ip_mask}.yaml",
    unless  => "${::calico::helper_dir}/calico_helper.sh get ${::calico::config_dir}/ipPool-${ip_pool}-${ip_mask}.yaml | /usr/bin/grep ${ip_pool}/${ip_mask}",
    require => [ Class['calico::node'], File["${::calico::install_dir}/bin/calicoctl"] ],
  }
}
