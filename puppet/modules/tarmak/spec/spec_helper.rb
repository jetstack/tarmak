require 'puppetlabs_spec_helper/module_spec_helper'

RSpec.configure do |config|
  config.default_facts = {
    :path => '/bin:/sbin:/usr/bin:/usr/sbin:/opt/bin',
    :ipaddress => '10.10.10.10',
    :osfamily => 'RedHat',
    :operatingsystemrelease => "7.5",
  }
end
