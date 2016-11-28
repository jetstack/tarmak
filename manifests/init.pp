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
  file { ['/etc/cni/net.d', '/opt/cni/bin']:
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
    mode => '755',
  }
  wget::fetch { "calico-ipam-v${calico_cni_version}":
    source => "https://github.com/projectcalico/calico-cni/releases/download/v${calico_cni_version}/calico-ipam",
    destination => '/opt/cni/bin/',
    mode => '755',
  }
}

define calico::lo_install (
  String $cni_plugin_version,
)
{
  archive { "download and extract cni-lo version $cni_plugin_version":
    source => "https://github.com/containernetworking/cni/releases/download/v${cni_plugin_version}/cni-v${cni_plugin_version}.tgz",
    extract_path => '/opt/cni/bin/',
    extract_flags => '-xzf loopback',
    creates => '/opt/cni/bin/loopback',
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
