# The baseline for module testing used by Puppet Labs is that each manifest
# should have a corresponding test manifest that declares that class or defined
# type.
#
# Tests are then run by using puppet apply --noop (to check for compilation
# errors and view a log of events) or by fully applying the test in a virtual
# environment (to compare the resulting system state to the desired state).
#
# Learn more about module testing here:
# https://docs.puppet.com/guides/tests_smoke.html
#

$etcd_cluster = [ 'etcd1' ]

class calico_install {
  class { 'calico':
    etcd_cluster => $etcd_cluster,
  }
}

class calico_policy_controller {
  class { 'calico::policy_controller': }
}

class calico_node_ip_pool_1 {
  calico::ip_pool { '10.234.0.0/16':
    ip_pool      => '10.234.0.0',
    ip_mask      => 16,
    ipip_enabled => 'true', #lint:ignore:quoted_booleans
  }
}

