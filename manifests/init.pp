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
  
  exec { 'Disable source dest check':
    command => @(CMD)'/bin/bash -c "REGION=$(/usr/bin/curl -s http://169.254.169.254/latest/dynamic/instance-identity/document | /usr/bin/jq -r .region); \
      INSTANCE=$(/usr/bin/curl -s http://169.254.169.254/latest/dynamic/instance-identity/document | /usr/bin/jq -r .instanceId); \
      /usr/bin/aws ec2 modify-instance-attribute --instance-id $INSTANCE --no-source-dest-check --region $REGION"'
      | CMD
      ,
    unless  => @(UNL)'/bin/bash -c "REGION=$(/usr/bin/curl -s http://169.254.169.254/latest/dynamic/instance-identity/document | /usr/bin/jq -r .region); \
      INSTANCE=$(/usr/bin/curl -s http://169.254.169.254/latest/dynamic/instance-identity/document | /usr/bin/jq -r .instanceId); \
      /usr/bin/aws ec2 describe-instance-attribute --instance-id $INSTANCE --attribute sourceDestCheck --region $REGION | /usr/bin/jq .SourceDestCheck.Value | /usr/bin/grep false"'
      | UNL
      ,
  }
}

class calico::policy_controller {
  file { "/root/calico-config.yaml":
    ensure => file,
    content => template('calico/calico-config.yaml.erb'),
  }

  file { "/root/policy-controller-deployment.yaml":
    ensure => file,
    content => template('calico/policy-controller-deployment.yaml.erb'),
  }

  exec { "deploy calico config":
    command => "/usr/bin/kubectl apply -f /root/calico-config.yaml",
    unless => "/usr/bin/kubectl get -f /root/calico-config.yaml",
    require => File["/root/calico-config.yaml"],
  }
  
  exec { "deploy calico policy controller":
    command => "/usr/bin/kubectl apply -f /root/policy-controller-deployment.yaml",
    unless => "/usr/bin/kubectl get -f /root/policy-controller-deployment.yaml",
    require => [ Exec["deploy calico config"], File["/root/policy-controller-deployment.yaml"], Exec["Trigger etcd overlay cert"] ],
  } 
}
