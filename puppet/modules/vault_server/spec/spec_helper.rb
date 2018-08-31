require 'puppetlabs_spec_helper/module_spec_helper'

RSpec.configure do |config|
  config.default_facts = {
    :osfamily => 'RedHat',
    :path => '/bin:/sbin:/usr/bin:/usr/sbin',
    :disks => {},
    :consul_master_token => '',
    :consul_encrypt => '',
    :consul_bootstrap_expect => '1',
  }
end
