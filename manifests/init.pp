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

  file { '/usr/local/bin/sourcedestcheck.sh':
    ensure  => file,
    content => template('calico/sourcedestcheck.sh.erb'),
    mode    => '0755',
  }

  exec { 'Disable source dest check':
    command => '/usr/local/bin/sourcedestcheck.sh set',
    unless  => '/usr/local/bin/sourcedestcheck.sh test',
    require => File['/usr/local/bin/sourcedestcheck.sh'],
  }
}
