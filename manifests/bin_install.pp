define calico::bin_install (
  String $calico_cni_version,
)
{
  wget::fetch { "calico-v${calico_cni_version}":
    source      => "https://github.com/projectcalico/calico-cni/releases/download/v${calico_cni_version}/calico",
    destination => '/opt/cni/bin/',
    require     => Class['calico'],
    before      => File['/opt/cni/bin/calico']
  }

  wget::fetch { "calico-ipam-v${calico_cni_version}":
    source      => "https://github.com/projectcalico/calico-cni/releases/download/v${calico_cni_version}/calico-ipam",
    destination => '/opt/cni/bin/',
    require     => Class['calico'],
    before      => File['/opt/cni/bin/calico-ipam'],
  }

  file { ['/opt/cni/bin/calico','/opt/cni/bin/calico-ipam']:
    ensure => file,
    mode   => '0755',
  }
}
