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
    require => Class['calico'],
    before => File["/opt/cni/bin/calico"]
  }
  
  wget::fetch { "calico-ipam-v${calico_cni_version}":
    source => "https://github.com/projectcalico/calico-cni/releases/download/v${calico_cni_version}/calico-ipam",
    destination => '/opt/cni/bin/',
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
  include k8s
  
  wget::fetch { "calicoctl-v${calico_cni_version}":
    source => "https://github.com/projectcalico/calico-containers/releases/download/v${calico_node_version}/calicoctl",
    destination => '/opt/cni/bin/',
    require => Class['calico'],
    before => File["/opt/cni/bin/calicoctl"],
  }

  file { ["/opt/cni/bin/calicoctl"]:
    ensure => file,
    mode   => '0755',
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
    require => [ Class["k8s"], File["/etc/calico/calico.env"], File["/usr/lib/systemd/system/calico-node.service"], Exec["Trigger etcd overlay cert"] ],
  }
}

define calico::ipPool (
  String $ip_pool,
  Integer $ip_mask,
  String $ipip_enabled
)
{
  file { "/etc/calico/ipPool-${ip_pool}.yaml":
    ensure => file,
    content => template('calico/ipPool.yaml.erb'),
  }
  
  exec { "Configure calico ipPool for CIDR $ip_pool":
    path => [ '/bin', '/usr/bin' ],
    user => "root",
    command => "/bin/bash -c \"`/usr/bin/grep ETCD_ENDPOINTS /etc/calico/calico.env` `/usr/bin/grep ETCD_CERT_FILE /etc/calico/calico.env` `/usr/bin/grep ETCD_KEY_FILE /etc/calico/calico.env` `/usr/bin/grep ETCD_CA_CERT_FILE /etc/calico/calico.env` /opt/cni/bin/calicoctl apply -f /etc/calico/ipPool-${ip_pool}.yaml\"",
    unless => "/bin/bash -c \"`/usr/bin/grep ETCD_ENDPOINTS /etc/calico/calico.env` `/usr/bin/grep ETCD_CERT_FILE /etc/calico/calico.env` `/usr/bin/grep ETCD_KEY_FILE /etc/calico/calico.env` `/usr/bin/grep ETCD_CA_CERT_FILE /etc/calico/calico.env` /opt/cni/bin/calicoctl get -f /etc/calico/ipPool-${ip_pool}.yaml | /usr/bin/grep ${ip_pool}/${ip_mask}\"",
    require => [ Service["calico-node"], File["/opt/cni/bin/calicoctl"], File["/etc/calico/ipPool-${ip_pool}.yaml"], Exec["Trigger etcd overlay cert"] ],
  }
}

define calico::policy_controller (
  String $dns_root,
  Integer $etcd_count,
  Integer $calico_etcd_port
)
{
  file { "/root/calico-config.yaml":
    ensure => file,
    content => template('calico/calico-config.yaml.erb'),
  }

  file { "/root/policy-controller-rs.yaml":
    ensure => file,
    content => template('calico/policy-controller-rs.yaml.erb'),
  }

  exec { "deploy calico config":
    command => "/usr/bin/kubectl apply -f /root/calico-config.yaml",
    unless => "/usr/bin/kubectl get -f /root/calico-config.yaml",
    require => File["/root/calico-config.yaml"],
  }
  
  exec { "deploy calico policy controller":
    command => "/usr/bin/kubectl apply -f /root/policy-controller-rs.yaml",
    unless => "/usr/bin/kubectl get -f /root/policy-controller-rs.yaml",
    require => [ Exec["deploy calico config"], File["/root/policy-controller-rs.yaml"], Exec["Trigger etcd overlay cert"] ],
  } 
}
