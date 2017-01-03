class calico::config
{

  include ::calico

  file { "${::calico::cni_base_dir}/net.d/10-calico.conf":
    ensure  => file,
    content => template('calico/10-calico.conf.erb'),
    require => Class['calico'],
  }
}
