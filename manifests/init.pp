# Class: etcd
# ===========================
#
# Full description of class etcd here.
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
#    class { 'etcd':
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

class etcd
{
  user { 'etcd':
    ensure => present,
    uid => 873,
    gid => 873,
    shell => '/sbin/nologin',
    home => '/var/lib/etcd',
  }
}


define etcd::install (
  String $etcd_version,
) 
{
  wget::fetch { "download etcd version $etcd_version":
    source => "https://github.com/coreos/etcd/releases/download/v${etcd_version}/etcd-v${etcd_version}-linux-amd64.tar.gz",
    destination => '/root/',
    before => Exec["untar etcd version $etcd_version"],
  }
    
  exec { "untar etcd version $etcd_version":
      command => "/bin/tar -xvzf /root/etcd-v${etcd_version}-linux-amd64.tar.gz -C /root/",
      creates => "/root/etcd-v${etcd_version}-linux-amd64/etcd",
  }

  file { "install etcd version $etcd_version":
    path => "/bin/etcd-${etcd_version}",
    source => "/root/etcd-v${etcd_version}-linux-amd64/etcd",
    mode => '755',
    require => Exec["untar etcd version $etcd_version"],
  }

  file { "install etcdctl version $etcd_version":
    path => "/bin/etcdctl-${etcd_version}",
    source => "/root/etcd-v${etcd_version}-linux-amd64/etcdctl",
    mode => '755',
    require => Exec["untar etcd version $etcd_version"],
  }
}

define etcd::config (
  String $cluster_name,
  String $etcd_version,
  Integer $client_port,
  Integer $peer_port,
)
{
  file { "/etc/etcd/etcd-${cluster_name}.conf":
    ensure => file,
    content => template('etcd/etcd.conf.erb'),
  }
}

define etcd::systemd (
  String $cluster_name,
  String $etcd_version,
)
{
  file { "/usr/lib/systemd/system/etcd-${cluster_name}.service":
    ensure => file,
    content => template('etcd/etcd.service.erb'),
  }
}
