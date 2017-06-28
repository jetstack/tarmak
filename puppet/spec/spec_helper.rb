require 'puppetlabs_spec_helper/module_spec_helper'

RSpec.configure do |c|
  base_path = File.dirname(__FILE__)
  c.hiera_config = File.join(base_path, 'fixtures/hiera.yaml')
  c.module_path = "#{File.join(File.dirname(base_path), 'modules')}:#{File.join(File.dirname(base_path), 'spec/fixtures/modules')}"
  c.manifest = File.join(File.dirname(base_path), 'manifests/site.pp')
  c.default_facts = {
    :tarmak_environment => 'nonprod',
    :tarmak_cluster => 'cluster1',
    :tarmak_dns_root => 'domain-zone.root',
    :path => '/usr/local/bin:/usr/bin:/usr/local/sbin:/usr/sbin:/opt/puppetlabs/bin',
    :ec2_metadata => {
      'instance_id' => 'i-fake',
      'placement' => {
        'availability-zone' => 'eu-west-1a',
      },
    },
    :ipaddress => '1.2.3.4',
    :osfamily => 'RedHat',
    :vault_token => 'init-token1',
  }
end
