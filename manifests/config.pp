class calico::config
{
  file { "${::calico::cni_base_dir}/cni/net.d/10-calico.conf":
    ensure  => file,
    content => template('calico/10-calico.conf.erb'),
    require => Class['calico'],
  }
}
