require 'puppetlabs_spec_helper/module_spec_helper'

RSpec.configure do |config|
  config.default_facts = {
    :osfamily => 'RedHat',
    :disks => {},
  }
end
