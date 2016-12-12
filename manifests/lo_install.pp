define calico::lo_install (
  String $cni_plugin_version,
)
{
  archive { "download and extract cni-lo version $cni_plugin_version":
    source => "https://github.com/containernetworking/cni/releases/download/v${cni_plugin_version}/cni-v${cni_plugin_version}.tgz",
    path => "/tmp/cni-v${cni_plugin_version}.tgz",
    extract => true,
    extract_path => '/opt/cni/bin/',
    extract_command => 'tar -xzf %s ./loopback',
    creates => '/opt/cni/bin/loopback',
    require => Class['calico'],
  }
}

