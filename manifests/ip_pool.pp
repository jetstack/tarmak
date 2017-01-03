define calico::ip_pool (
  String $ip_pool,
  Integer $ip_mask,
  String $ipip_enabled
)
{

  include ::calico

  file { "${::calico::helper_dir}/calico_helper.sh":
    ensure  => file,
    content => template('calico/calico_helper.sh.erb'),
    mode    => '0755',
  } ->
  file { "/etc/calico/ipPool-${ip_pool}.yaml":
    ensure  => file,
    content => template('calico/ipPool.yaml.erb'),
  } ->
  exec { "Configure calico ipPool for CIDR ${ip_pool}":
    user    => 'root',
    command => "${::calico::helper_dir}/calico_helper.sh apply -f ${::calico::config_dir}/ipPool-${ip_pool}.yaml",
    unless  => "${::calico::helper_dir}/calico_helper.sh get -f ${::calico::config_dir}/ipPool-${ip_pool}.yaml | /usr/bin/grep ${ip_pool}/${ip_mask}",
    require => [ Service['calico-node'], File["${::calico::install_dir}/bin/calicoctl"] ],
  }
}
