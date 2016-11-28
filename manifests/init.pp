# Class: calico
# ===========================
#
# Full description of class calico here.
#
# Parameters
# ----------
#
# Document parameters here.
#
# * `sample parameter`
# Explanation of what this parameter affects and what it defaults to.
# e.g. "Specify one or more upstream ntp servers as an array."
#
# Variables
# ----------
#
# Here you should define a list of variables that this module would require.
#
# * `sample variable`
#  Explanation of how this variable affects the function of this class and if
#  it has a default. e.g. "The parameter enc_ntp_servers must be set by the
#  External Node Classifier as a comma separated list of hostnames." (Note,
#  global variables should be avoided in favor of class parameters as
#  of Puppet 2.6.)
#
# Examples
# --------
#
# @example
#    class { 'calico':
#      servers => [ 'pool.ntp.org', 'ntp.local.company.com' ],
#    }
#
# Authors
# -------
#
# Author Name <author@domain.com>
#
# Copyright
# ---------
#
# Copyright 2016 Your name here, unless otherwise noted.
#

class calico
{
  file { ['/etc/cni', '/etc/cni/net.d', '/etc/calico', '/opt/cni', '/opt/cni/bin']:
    ensure => directory,
  }
}

define calico::bin_install (
  String $calico_cni_version,
)
{
  wget::fetch { "calico-v${calico_cni_version}":
    source => "https://github.com/projectcalico/calico-cni/releases/download/v${calico_cni_version}/calico",
    destination => '/opt/cni/bin/',
    mode => '0755',
    require => Class['calico'],
    before => File["/opt/cni/bin/calico"]
  }
  wget::fetch { "calico-ipam-v${calico_cni_version}":
    source => "https://github.com/projectcalico/calico-cni/releases/download/v${calico_cni_version}/calico-ipam",
    destination => '/opt/cni/bin/',
    mode => '0755',
    require => Class['calico'],
    before => File["/opt/cni/bin/calico-ipam"],
  }
  file { ["/opt/cni/bin/calico","/opt/cni/bin/calico-ipam"]:
    ensure => file,
    mode   => '0755',
  }
  
}

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


define calico::config (
  String $calico_name,
  Integer $etcd_count,
  Integer $calico_etcd_port,
)
{
  file { "/etc/cni/net.d/10-calico.conf":
    ensure => file,
    content => template('calico/10-calico.conf.erb'),
    require => Class['calico'],
  }
}

define calico::node (
  String  $calico_node_version,
  Integer $etcd_count,
  Integer $calico_etcd_port,
)

{
  include ::systemd

  package { "docker":
    ensure => installed,
  }

  file { "/etc/calico/calico.env":
    ensure => file,
    content => template('calico/calico.env.erb'),
    require => Class['calico'],
  }

  file { "/usr/lib/systemd/system/calico-node.service":
    ensure => file,
    content => template('calico/calico-node.service.erb'),
  } ~>
  Exec['systemctl-daemon-reload']
  
  service { "calico-node":
    ensure => running,
    enable => true,
    require => [ Package["docker"], File["/etc/calico/calico.env"], File["/usr/lib/systemd/system/calico-node.service"] ],
  }
}
